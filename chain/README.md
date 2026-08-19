# Chain

is a tool for chaining other tools:

```
chain --storage ./path + progA arg + progB arg1 arg2 + progC
```

Default for `--storage` is `.`. First positional argument to `chain` is the subprogram separator

1. It will remove every `*.socket` file in the storage directory
2. It will call `progC`, passing `--input ./path/progC_3.socket` to it before other arguments (template: `{storage directory}{OS path separator}{program file name}_{subprogram number}.socket`)
3. It will wait until `progC`'s `--input` file gets created by `progC`
4. It will call `progB`, passing `--input ./path/progB_2.socket` and `--output ./path/progC_3.socket`
5. It will wait for `./path/progB_2.socket` to exist
6. It will call `progA`, passing `--output ./path/progB_2.socket` to it

This way, programs write their outputs into the next program and read their inputs from the previous program. The first program doesn't have an input socket, the last program doesn't have an output socket

Subprograms are ran until one of them exits or a termination signal comes into `chain` (for example, pressing `ctrl-c` in a terminal), at which point all get killed

Stdout and stderr from every program get `{program name}_{subprogram number}` prepended before them, like `progB_2 actual log text`
