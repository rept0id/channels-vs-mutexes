# channels-vs-mutexes

![header](./doc/assets/img/header.png)

Wow !

Sometimes I think Go is literally the 21st century’s C.

Here is a comparison of two implementations: one using the classic mutual exclusions (mutexes) -that you find in any language- and the other using Go's channels. Both implementations have two goroutines (Go's "threads"), the first prints "red" and the second prints "blue" over a connection.

The implementation using Go's channels results in a more natural and even switch between "red" and "blue". You see much less "red, red, blue, blue ..." and much more "red, blue, red, blue ...".

This is simply because channels coordinate access to resources, avoiding the need to fight for them as mutual exclusions require.

For anyone interested further, you can read about "Communicating Sequential Processes" (CSP) by Tony Hoare and the "share memory by communicating" principle.
