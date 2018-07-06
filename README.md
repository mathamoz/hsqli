Work in progress golang utility to save combined bash history from all terminal sessions into a sqlite database

NB: Make sure to save off a copy of your `~/.bash_history` before using this tool, otherwise you will lose it!  Also, this is a very quick and rough utility that needs some more work but it should be stable as-is, though your mileage may vary.

Place the `hsqli` binary somewhere in your path and then add the following to your `~/.bashrc`

```
prompt_command() {
    history | sed '$!d' | cut -c 8- | hsqli && history -c && hsqli -fetch && history -r
}

PROMPT_COMMAND=prompt_command
```

Source your new `~/.bashrc` or reload your terminal session(s) and everything should work.
