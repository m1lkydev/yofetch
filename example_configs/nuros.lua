--[[
    Remake of NurOS (the FRESH operating system) Yofetch Configuration by m1lkydev.
    The original can be found at https://github.com/NurOS-Linux/nuros-yofetch-config.
]]

--[[ 
    NurOS is an independent Linux distribution 
    designed for high speed, security, and accessibility.
    Our goal is to create an open and powerful operating system 
    for developers and everyday users.

    For more information, visit https://nuros.org.
]]

local yo = require("Yofetch")

local accent = "{style.bold}{color.blue}"
local reset = "{style.reset}"
local sep = reset .. "{color.cyan} - " .. reset
local user_host_sep = "-"

local function info_line(label, value)
    if label == "" then
        return accent .. value .. reset
    end
    return accent .. label .. sep .. reset .. value
end

local function slogan()
    return "\27[90m" .. "Shine brighter than the rest by thinking different" .. reset
end

local function parse_flag_value(flag)
    local pattern_with_value = "%-" .. flag .. "='(.-)'"
    local raw_value = yo.config_args:match(pattern_with_value)

    if raw_value then
        if raw_value:find(",") then
            local values = {}
            for val in raw_value:gmatch("([^,%s]+)") do
                table.insert(values, val)
            end
            return values
        else
            return raw_value
        end
    end

    local pattern_flag_only = "%-" .. flag .. "%s"
    if yo.config_args:find(pattern_flag_only) or yo.config_args:match("%-" .. flag .. "$") then
        return true
    end

    return false
end

yo.padding(3, 0, 0) -- Right, Left, Top
yo.mode("default")  -- Horizontal layout
yo.shell({ "/bin/sh", "-c" })

if parse_flag_value("braille") then
    yo.logo([[
{color.blue}⠀⠀⢀⣤⣶⠿⠿⠿⠿⢷⣶⣄⡀⠀⠀
{color.blue}⠀⣰⡿⠋⠁⢀⣠⣤⣀⡀⠈⠙⢿⣄⠀
{color.blue}⣸⡿⠁⢀⣾⣿⣿⣿⣿⣿⣷⡀⠈⢿⡆
{color.blue}⣿⠃⠀⣾⣿⣿⣿⣿⣿⣿⣿⣷⠀⠸⣿
{color.blue}⣿⡄⠀⢿⣿⣿⣿⣿⣿⣿⣿⡽⠀⢰⣟
{color.blue}⠸⣷⡀⠈⠿⣿⣿⣿⣿⣞⠷⠁⢀⣾⠇
{color.blue}⠀⠙⣷⣄⡀⠈⠉⠉⠁⠀⢀⣤⡿⠋⠀
{color.blue}⠀⠀⠀⠙⠿⢷⣶⣶⣶⠾⠟⠋⠀⠀⠀
]])
else
    yo.logo([[
{color.blue}         ___╓╓___  
{color.blue}     _▄▄▓▓▀▀╜╜╨▀▓▓▓╗_ 
{color.blue}   ╓▓▓▀²          `╙▓▓╖ 
{color.blue}  ╣▓▀   _▄▓▓▓▓▓▓W_   ╙▓▓ 
{color.blue} ╣▓╜  ,▓▓▓▓▓▓▓▓▓▓▓▓_  ²▓▓
{color.blue}╒▓▌   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ║▓m
{color.blue}╞▓▓  í▓▓▓▓▓▓▓▓▓▓▓▓▓▓h  ╞▓╡
{color.blue}²▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ║▓h
{color.blue} ║▓▄   ╙▓▓▓▓▓▓▓▓▓▓╜   ƒ▓▓
{color.blue}  ╙▓▓_   ⁿ╙╨╝╝╝╜²   _╢▓╜
{color.blue}    ╙▓▓╗__       _╗▓▓╜
{color.blue}      `╙╝▓▓▓▓▓▓▓▓╝╙
]])
end

local user_host = yo.exec("echo $(whoami)@$(uname -n)")
local kernel = yo.exec("echo $(uname -s) $(uname -r) $(uname -m)")
local uptime = yo.exec("uptime -p | sed 's/up //'")

local packages = (function()
    local tulpar_count = yo.exec("command -v tulpar >/dev/null 2>&1 && tulpar list --count 2>/dev/null || echo ''")
    if tulpar_count ~= "" and tonumber(tulpar_count) and tonumber(tulpar_count) > 0 then
        return tulpar_count .. " (tulpar)"
    end

    local managers = {
        { "pacman -Qq 2>/dev/null | wc -l",           "pacman" },
        { "dpkg -l 2>/dev/null | grep '^ii' | wc -l", "apt" },
        { "rpm -qa 2>/dev/null | wc -l",              "dnf" },
        { "nix-env -q 2>/dev/null | wc -l",           "nix" },
        { "flatpak list 2>/dev/null | wc -l",         "flatpak" }
    }

    for _, manager in ipairs(managers) do
        local count = yo.exec(manager[1])
        local num = tonumber(count)
        if num and num > 0 then
            return count .. " (" .. manager[2] .. ")"
        end
    end

    return "unknown"
end)()

local shell = yo.exec("basename $SHELL")

local de = (function()
    local de = yo.exec("echo $XDG_CURRENT_DESKTOP")
    if de == "" then
        de = yo.exec("echo $DESKTOP_SESSION")
    end
    if de == "" then
        de = yo.exec("wmctrl -m 2>/dev/null | grep 'Name:' | cut -d' ' -f2")
    end
    if de == "" then
        de = "TTY"
    end
    return de
end)()

local terminal = (function()
    local methods = {
        "echo $TERMINAL",
        "ps -o comm= $(ps -o ppid= $(ps -o ppid= $(ps -o ppid= $$)))",
        "pstree -s $$ | grep -oE '(kitty|alacritty|konsole|gnome-terminal|terminator|tilix|xterm|urxvt|st|foot|wezterm|rio)' | head -1",
        "cat /proc/$(ps -o ppid= $(ps -o ppid= $$))/comm 2>/dev/null"
    }

    for _, cmd in ipairs(methods) do
        local term = yo.exec(cmd)
        if term ~= "" and term ~= "yofetch" and term ~= "bash" and term ~= "zsh" and term ~= "fish" then
            return term
        end
    end

    return yo.exec("echo $TERM")
end)()

yo.print(info_line("", user_host))
yo.print(info_line("", reset .. string.rep(user_host_sep, #user_host)))
yo.print(info_line("OS", "NurOS"))
yo.print(info_line("Kernel", kernel))
yo.print(info_line("Uptime", uptime))
yo.print(info_line("Packages", packages))
yo.print(info_line("Shell", shell))
yo.print(info_line("DE", de))
yo.print(info_line("Terminal", terminal))
yo.print("")

if parse_flag_value("slogan") == true then
    yo.print(slogan())
    yo.print("")
end

yo.print(
    "{color.black}███{color.red}███{color.green}███{color.yellow}███{color.blue}███{color.magenta}███{color.cyan}███{color.white}███{style.reset}\n"..
    "{color.black_light}███{color.red_light}███{color.green_light}███{color.yellow_light}" ..
    "███{color.blue_light}███{color.magenta_light}███{color.cyan_light}███{color.white_light}███{style.reset}\n"
)
