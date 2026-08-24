# Chain

is a tool for chaining other tools:

```
chain --storage ./path + progA arg + progB arg1 arg2 + progC
```

Default for `--storage` is `.`. First positional argument to `chain` is the subprogram separator

1. It will create the `--storage` directory if it doesn't exist, including all the parent directories
2. It will remove every `*.socket` file in the storage directory
3. It will call `progC`, passing environment variable `INPUT=./path/progC_3.socket` to it (template: `{storage directory}{OS path separator}{program file name}_{subprogram number}.socket`)
4. It will wait until `progC`'s `INPUT` file gets created by `progC`
5. It will call `progB`, passing `INPUT=./path/progB_2.socket` and `OUTPUT=./path/progC_3.socket`
6. It will wait for `./path/progB_2.socket` to exist
7. It will call `progA`, passing `OUTPUT=./path/progB_2.socket` to it

This way, programs write their outputs into the next program and read their inputs from the previous program. The first program doesn't have an input socket, the last program doesn't have an output socket

Subprograms are ran until one of them exits or a termination signal comes into `chain` (for example, pressing `ctrl-c` in a terminal), at which point all get killed

Stdout and stderr from every program get `{program name}_{subprogram number}:` prepended before them, like `progB_2: actual log text`
