# Yofetch Lua config info

The `Yofetch` module contains a:

## Functions
| Function name       | Value type           | Description                                  |
|---------------------|----------------------|----------------------------------------------|
| `print(text)`       | Lua string           | Printing the text in information             |
| `logo(logo)`        | Any Lua string       | Sets the logo                                |
| `padding(r, l, t)`  | Lua int              | Sets indents: right, left, top               |
| `mode(value)`       | Lua string           | Sets the mode                                |
| `shell({...})`      | Lua table            | Sets the path to the shell and its arguments |
| `exec(command)`     | Any Lua string       | Executes the command and returns the result  |

## Variables
| Variable name        | Value type       | Description                               |
|----------------------|------------------|-------------------------------------------|
| `version`            | Lua string       | Stores a version of Yofetch               |
| `config_args`        | Lua string       | Config arguments passed via `-a` CLI flag |

## Colors
| Color ID                |
|-------------------------|
| `{color.black}`         |
| `{color.red}`           |
| `{color.green}`         |
| `{color.yellow}`        |
| `{color.blue}`          |
| `{color.magenta}`       |
| `{color.cyan}`          |
| `{color.white}`         |
| `{color.black_light}`   |
| `{color.red_light}`     |
| `{color.green_light}`   |
| `{color.yellow_light}`  |
| `{color.blue_light}`    |
| `{color.magenta_light}` |
| `{color.cyan_light}`    |
| `{color.white_light}`   |

## Styles
| Style ID                   |
|----------------------------|
| `{style.reset}`            |
| `{style.bold}`             |
| `{style.faint}`            |
| `{style.italic}`           |
| `{style.underline}`        |
| `{style.blink}`            |
| `{style.blink_fast}`       |
| `{style.reverse}`          |
| `{style.hidden}`           |
| `{style.strikethrough}`    |
| `{style.double_underline}` |
| `{style.overline}`         |


# How to create config
1) Create a file with a .lua extension in any folder (for example, `config.lua` in the root folder with yofetch)
2) Import the `Yofetch` module in the config:
```lua
local yo = require("Yofetch")
```
3) Use the available functions/variables through the imported `yo`:
```lua
yo.padding(3, 0, 0)
yo.mode("default")
yo.shell({"/bin/sh", "-c"})

yo.print("{style.italic}{color.red}" .. "OS{style.reset}: " .. yo.exec(". /etc/os-release; echo $PRETTY_NAME $(uname -m)"))
yo.print("{style.italic}{color.red}" .. "Kernel{style.reset}: " .. yo.exec("echo $(uname -r)"))
yo.print("{style.italic}{color.red}" .. "Uptime{style.reset}: " .. yo.exec("echo $(uptime -p | sed 's/^up //')"))
yo.print("{style.italic}{color.red}" .. "Mem{style.reset}: " .. yo.exec([[
    free -k | awk '/Mem:/ {printf "%.1f GiB / %.1f GiB\n", $3/1048576, $2/1048576}'
]]))
```

This is just a sample config, you can supplement it with your imagination.