# Trasher

accepts byte streams on an `INPUT` (environment variable) Unix socket and makes respective connections to the `OUTPUT` (environment variable) Unix socket.

Everything in the `$INPUT` -> `$OUTPUT` direction is "trashed": extra bytes are added to streams in random places in a way that will be undoable later

The `$OUTPUT` -> `$INPUT` direction untrashes a trashed stream

A `--clean` flag makes it clean `$INPUT` -> `$OUTPUT` and trash `$OUTPUT` -> `$INPUT`

About the trashing direction:
* For each incoming byte, there are about `--trash-ratio` bytes, so if `--trash-ratio` is `0.3`, about 30% of outgoing bytes are trash. The default for `--trash-ratio` is `0.3`
* Trash bytes are distributed somewhat evenly among source bytes, so on average, the `--trash-ratio` is true for any part of the connection
* The trashing patterns are always random for every stream, so two identical input streams produce different output streams (unless extremely unlucky)
