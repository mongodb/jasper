package jasper

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
)

const (
	// Constants for error codes.
	// Documentation: https://docs.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-

	errSuccess          syscall.Errno = 0
	errAccessDenied     syscall.Errno = 5
	errInvalidParameter syscall.Errno = 87

	// Constants for process and standard access rights.
	// Documentation: https://doc.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights

	standardRightRequired     = 0x000F0000
	standardRightSynchronize  = 0x00100000
	procRightTerminate        = 0x0001
	procRightQueryInformation = 0x0400
	procRightAllAccess        = standardRightRequired | standardRightSynchronize | 0xFFFF

	// Constants for additional configuration for job object API calls.
	// Documentation: https://docs.microsoft.com/en-us/windows/win32/procthread/job-objects

	jobObjectLimitKillOnJobClose = 0x2000

	jobObjectInfoClassNameBasicProcessIDList       = 3
	jobObjectInfoClassNameExtendedLimitInformation = 9

	// Constants for process exit codes.

	procStillActive = 0x103

	// Constants for event object access rights.

	eventRightModifyState = 0x0002

	// Constants representing the wait return value.

	waitFailed uint32 = 0xFFFFFFFF

	// Max allowed length of the process ID list.
	maxProcessIDListLength = 1000
)

var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procAssignProcessToJobObject  = modkernel32.NewProc("AssignProcessToJobObject")
	procCloseHandle               = modkernel32.NewProc("CloseHandle")
	procCreateJobObjectW          = modkernel32.NewProc("CreateJobObjectW")
	procCreateEventW              = modkernel32.NewProc("CreateEventW")
	procOpenEventW                = modkernel32.NewProc("OpenEventW")
	procSetEvent                  = modkernel32.NewProc("SetEvent")
	procResetEvent                = modkernel32.NewProc("ResetEvent")
	procGetExitCodeProcess        = modkernel32.NewProc("GetExitCodeProcess")
	procOpenProcess               = modkernel32.NewProc("OpenProcess")
	procTerminateProcess          = modkernel32.NewProc("TerminateProcess")
	procQueryInformationJobObject = modkernel32.NewProc("QueryInformationJobObject")
	procTerminateJobObject        = modkernel32.NewProc("TerminateJobObject")
	procSetInformationJobObject   = modkernel32.NewProc("SetInformationJobObject")
	procWaitForSingleObject       = modkernel32.NewProc("WaitForSingleObject")
)

// JobObject is a wrapper for a Windows job object.
type JobObject struct {
	handle syscall.Handle
}

// NewWindowsJobObject creates a new Windows object object with the given name.
func NewWindowsJobObject(name string) (*JobObject, error) {
	utf16Name, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, NewWindowsError("UTF16PtrFromString", err)
	}
	handle, err := CreateJobObject(utf16Name)
	if err != nil {
		return nil, NewWindowsError("CreateJobObject", err)
	}

	if err := SetInformationJobObjectExtended(handle, JobObjectExtendedLimitInformation{
		BasicLimitInformation: JobObjectBasicLimitInformation{
			LimitFlags: jobObjectLimitKillOnJobClose,
		},
	}); err != nil {
		return nil, NewWindowsError("SetInformationJobObject", err)
	}

	return &JobObject{handle: handle}, nil
}

// AssignProcess assigns a process to this job object.
func (j *JobObject) AssignProcess(pid uint) error {
	hProcess, err := OpenProcess(procRightAllAccess, false, uint32(pid))
	if err != nil {
		return NewWindowsError("OpenProcess", err)
	}
	defer func() {
		if err := NewWindowsError("CloseHandle", CloseHandle(hProcess)); err != nil {
			grip.Warning(message.WrapError(err, message.Fields{
				"message": "failed to close job object handle",
				"pid":     pid,
				"op":      "AssignProcess",
			}))
		}
	}()

	if err := AssignProcessToJobObject(j.handle, hProcess); err != nil {
		return NewWindowsError("AssignProcessToJobObject", err)
	}
	return nil
}

// Terminate terminates all processes associated with this job object.
func (j *JobObject) Terminate(exitCode uint) error {
	if err := TerminateJobObject(j.handle, uint32(exitCode)); err != nil {
		return NewWindowsError("TerminateJobObject", err)
	}
	return nil
}

// Close closes the job object's handle.
func (j *JobObject) Close() error {
	if j.handle != 0 {
		if err := CloseHandle(j.handle); err != nil {
			return NewWindowsError("CloseHandle", err)
		}
		j.handle = 0
	}
	return nil
}

// Event is a wrapper for a Windows event object.
type Event struct {
	handle syscall.Handle
}

// NewEvent creates a new event object with the given name or opens an existing
// event with the given name.
func NewEvent(name string) (*Event, error) {
	utf16Name, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, NewWindowsError("UTF16PtrFromString", err)
	}
	handle, err := CreateEvent(utf16Name)
	if err != nil {
		return nil, NewWindowsError("CreateEvent", err)
	}
	return &Event{handle: handle}, nil
}

// GetEvent gets an existing event object with the given name.
func GetEvent(name string) (*Event, error) {
	utf16Name, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, NewWindowsError("UTF16PtrFromString", err)
	}
	handle, err := OpenEvent(utf16Name)
	if err != nil {
		return nil, NewWindowsError("OpenEvent", err)
	}
	return &Event{handle: handle}, nil
}

// Set sets the event object to the signaled state.
func (event *Event) Set() error {
	if err := SetEvent(event.handle); err != nil {
		return NewWindowsError("SetEvent", err)
	}
	return nil
}

// Reset resets the event object to the non-signaled state.
func (event *Event) Reset() error {
	if err := ResetEvent(event.handle); err != nil {
		return NewWindowsError("ResetEvent", err)
	}
	return nil
}

// Close closes the event object's handle.
func (event *Event) Close() error {
	if err := CloseHandle(event.handle); err != nil {
		return NewWindowsError("CloseHandle", err)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////
//
// All the methods below are boilerplate functions for accessing the Windows syscalls.
//
///////////////////////////////////////////////////////////////////////////////////////////

// IOCounters includes accounting information for a process or job object.
type IOCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

// JobObjectBasicProcessIDList returns information about the processes running
// in a job object.
type JobObjectBasicProcessIDList struct {
	NumberOfAssignedProcesses uint32
	NumberOfProcessIDsInList  uint32
	ProcessIDList             [maxProcessIDListLength]uint64
}

// JobObjectBasicLimitInformation contains basic information to configure a job
// object's limits.
type JobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit uint64
	PerJobUserTimeLimit     uint64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

// JobObjectExtendedLimitInformation contains extended information to configure
// a job object's limits.
type JobObjectExtendedLimitInformation struct {
	BasicLimitInformation JobObjectBasicLimitInformation
	IOInfo                IOCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

// SecurityAttributes specifies the security configuration for a job object.
type SecurityAttributes struct {
	Length             uint32
	SecurityDescriptor uintptr
	InheritHandle      uint32
}

// OpenProcess opens an existing process object.
func OpenProcess(desiredRight uint32, inheritHandle bool, pid uint32) (syscall.Handle, error) {
	var inheritHandleRaw int32
	if inheritHandle {
		inheritHandleRaw = 1
	}
	r1, _, e1 := procOpenProcess.Call(
		uintptr(desiredRight),
		uintptr(inheritHandleRaw),
		uintptr(pid))
	if r1 == 0 {
		if e1 != errSuccess {
			return 0, e1
		}
		return 0, syscall.EINVAL
	}
	return syscall.Handle(r1), nil
}

// GetExitCodeProcess gets the exit status of the process given by the handle.
func GetExitCodeProcess(handle syscall.Handle, exitCode *uint32) error {
	r1, _, e1 := procGetExitCodeProcess.Call(uintptr(handle), uintptr(unsafe.Pointer(exitCode)))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// TerminateProcess terminates the process given by the handle.
func TerminateProcess(handle syscall.Handle, exitCode uint32) error {
	r1, _, e1 := procTerminateProcess.Call(uintptr(handle), uintptr(exitCode))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// CreateJobObject creates a new job object with the given name.
func CreateJobObject(name *uint16) (syscall.Handle, error) {
	jobAttributes := &SecurityAttributes{}

	r1, _, e1 := procCreateJobObjectW.Call(
		uintptr(unsafe.Pointer(jobAttributes)),
		uintptr(unsafe.Pointer(name)))
	if r1 == 0 {
		if e1 != errSuccess {
			return 0, e1
		}
		return 0, syscall.EINVAL
	}
	return syscall.Handle(r1), nil
}

// AssignProcessToJobObject assigns a process to a job object given by the
// handle.
func AssignProcessToJobObject(job syscall.Handle, process syscall.Handle) error {
	r1, _, e1 := procAssignProcessToJobObject.Call(uintptr(job), uintptr(process))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// QueryInformationJobObjectProcessIDList gets information about the processes
// in a job object.
func QueryInformationJobObjectProcessIDList(job syscall.Handle) (*JobObjectBasicProcessIDList, error) {
	var info JobObjectBasicProcessIDList
	_, err := QueryInformationJobObject(
		job,
		jobObjectInfoClassNameBasicProcessIDList,
		unsafe.Pointer(&info),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// QueryInformationJobObject gets information about a job object given by the
// handle.
func QueryInformationJobObject(job syscall.Handle, infoClass uint32, info unsafe.Pointer, length uint32) (uint32, error) {
	var nLength uint32
	r1, _, e1 := procQueryInformationJobObject.Call(
		uintptr(job),
		uintptr(infoClass),
		uintptr(info),
		uintptr(length),
		uintptr(unsafe.Pointer(&nLength)),
	)
	if r1 == 0 {
		if e1 != errSuccess {
			return 0, e1
		}
		return 0, syscall.EINVAL
	}
	return nLength, nil
}

// SetInformationJobObjectExtended sets limits for a job object given by the
// handle.
func SetInformationJobObjectExtended(job syscall.Handle, info JobObjectExtendedLimitInformation) error {
	r1, _, e1 := procSetInformationJobObject.Call(uintptr(job), jobObjectInfoClassNameExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uintptr(uint32(unsafe.Sizeof(info))))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// TerminateJobObject terminates all processes currently associated with the
// job.
func TerminateJobObject(job syscall.Handle, exitCode uint32) error {
	r1, _, e1 := procTerminateJobObject.Call(uintptr(job), uintptr(exitCode))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// CreateEvent creates or opens a named event object.
func CreateEvent(name *uint16) (syscall.Handle, error) {
	r1, _, e1 := procCreateEventW.Call(
		uintptr(unsafe.Pointer(nil)),
		uintptr(1),
		uintptr(0),
		uintptr(unsafe.Pointer(name)),
	)
	handle := syscall.Handle(r1)
	if r1 == 0 {
		if e1 != errSuccess {
			return handle, e1
		}
		return handle, syscall.EINVAL
	}
	return handle, nil
}

// OpenEvent opens an existing named event.
func OpenEvent(name *uint16) (syscall.Handle, error) {
	r1, _, e1 := procOpenEventW.Call(
		uintptr(eventRightModifyState),
		uintptr(0),
		uintptr(unsafe.Pointer(name)),
	)
	handle := syscall.Handle(r1)
	if r1 == 0 {
		if e1 != errSuccess {
			return handle, e1
		}
		return handle, syscall.EINVAL
	}
	return handle, nil
}

// SetEvent sets the event to the signaled state.
func SetEvent(event syscall.Handle) error {
	r1, _, e1 := procSetEvent.Call(uintptr(event))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// ResetEvent sets the event to the non-signaled state.
func ResetEvent(event syscall.Handle) error {
	r1, _, e1 := procResetEvent.Call(uintptr(event))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// WaitForSingleObject waits until the specified object is signaled or the
// timeout elapses.
func WaitForSingleObject(object syscall.Handle, timeout time.Duration) (uint32, error) {
	timeoutMillis := int64(timeout * time.Millisecond)
	r1, _, e1 := procWaitForSingleObject.Call(
		uintptr(object),
		uintptr(uint32(timeoutMillis)),
	)
	waitStatus := uint32(r1)
	if waitStatus == waitFailed {
		if e1 != errSuccess {
			return waitStatus, e1
		}
		return waitStatus, syscall.EINVAL
	}
	return waitStatus, nil
}

// CloseHandle closes an open object handle.
func CloseHandle(object syscall.Handle) error {
	r1, _, e1 := procCloseHandle.Call(uintptr(object))
	if r1 == 0 {
		if e1 != errSuccess {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

// WindowsError represents a Windows API error.
type WindowsError struct {
	innerError   error
	functionName string
}

// NewWindowsError creates a new Windows API error.
func NewWindowsError(functionName string, err error) error {
	if err == nil {
		return nil
	}
	return &WindowsError{innerError: err, functionName: functionName}
}

// FunctionName returns the name of the Windows API function.
func (e *WindowsError) FunctionName() string {
	return e.functionName
}

// InnerError returns the raw error from the Windows API.
func (e *WindowsError) InnerError() error {
	return e.innerError
}

// Error returns the formatted Windows API error.
func (e *WindowsError) Error() string {
	return fmt.Sprintf("gowin32: %s failed: %v", e.functionName, e.innerError)
}
