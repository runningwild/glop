Glop (Game Library Of Power) is a fairly simple cross-platform game library.

- ai (DEPRECATED)
- gin - Input manager, simple interface that supports buttons, mouse wheels and mouse axes, and a way of describing key-combos.
- gos - Os-specific code, every supported operating system must be made to conform to the system.System interface.
- gui - Simple gui toolkit.  This code is not good and should probably be rewritten completely.
- memory - For doing manual memory management if you need to avoid the gc or run things on a 32-bit system because of go's gc issues.
- render - Render thread.
- sprite - Supports making sprites with flowcharts created by yEd.
- system - Describes the interface that all supported operating systems must conform to.  This is seperated from gos so that it can be tested more easily.
- util - Some basic algorithms useful in a lot of places.

If you have any questions, please let me know!  runningwild@gmail.com