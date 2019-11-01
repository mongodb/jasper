package wire

// Note: these tests are largely copied directly from the top level
// package into this package to avoid an import cycle.

// func TestMongowireProcess(t *testing.T) {
//     for mname, makep := range map[string]func(ctx context.Context, t *testing.T) jasper.Process{
//         "Basic": func(ctx context.Context, t *testing.T) jasper.Process {
//             port := testutil.GetNextPort()
//             mgr, err := jasper.NewSynchronizedManager(false)
//             require.NoError(t, err)
//             service, err := NewManagerService(mgr, "localhost", port)
//             require.NoError(t, err)
//
//         },
//         "Blocking": func(ctx context.Context, t *testing.T) jasper.Process {
//         },
//     } {
//
//     }
// }
