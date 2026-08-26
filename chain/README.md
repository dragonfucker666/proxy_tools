# Chain

is a tool for chaining other tools:

```
chain ./socket_storage_directory SEPARATOR progA arg SEPARATOR progB arg1 arg2 SEPARATOR progC
```

First two arguments are required

1. It will create the socket storage directory if it doesn't exist, including all the parent directories
2. It will remove every `*.socket` file in the storage directory
3. It will call `progC`, passing environment variable `INPUT=./socket_storage_directory/3.socket` to it ("3" = subprogram number)
4. It will wait until `progC`'s `INPUT` file gets created by `progC`
5. It will call `progB`, passing `INPUT=./socket_storage_directory/2.socket` and `OUTPUT=./socket_storage_directory/3.socket`
6. It will wait for `./socket_storage_directory/2.socket` to exist
7. It will call `progA`, passing `OUTPUT=./socket_storage_directory/2.socket` to it

This way, programs write their outputs into the next program and read their inputs from the previous program. The first program doesn't have an input socket, the last program doesn't have an output socket

Subprograms are ran until one of them exits or a termination signal comes into `chain` (for example, pressing `ctrl-c` in a terminal), at which point all get killed

Stdout and stderr of every subprogram get prefixed with subprogram number, like `2: actual log text`
