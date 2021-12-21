package main

// THIS FILE IS AUTOGENERATE DO NOT EDIT

var runtimeCode = `
-- {{{ Globals
json = require("goluago/encoding/json")
strings = require("goluago/strings")

word_sep_chars = "()[]{}\"\'\\/ "
word_sep_chars_with_punctuation = "()[]{}.\"\'\\/ _-"
word_chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-"

buffers = {}
keymaps = {}
settings = {}
commands = {}
root_window = nil
current_window = nil
keys_entered = key("")
_message = ""
_message_type = "info"


string.join = function(arr, sep)
  local ret = arr[1] or ""
  for i = 2, #arr, 1 do
    ret = ret..sep..arr[i]
  end
  return ret
end

string.contains_char = function(chars, c)
  for i = 1, #chars, 1 do
    if c == string.sub(chars, i, i) then
      return true
    end
  end
  return false
end

string.find_char_index = function(str, chars)
  for i = 1, #str, 1 do
    if string.contains_char(chars, string.sub(str, i, i)) then
      return i
    end
  end
  return 0
end

id_counter = 0
function next_id()
  id_counter = id_counter + 1
  return id_counter
end

function debug_value(value)
  screen_quit()
  if type(value) == "table" then
    print(json.marshal(value))
  else
    print(value)
  end
  os.exit(1)
end

function chain(a, b, c)
  return function(...)
    if a then a(...) end
    if b then b(...) end
    if c then c(...) end
  end
end

function message(message, message_type)
  -- get
  if not message then
    return _message
  end

  -- set
  _message = message
  if message_type then
    _message_type = message_type
  else
    _message_type = "info"
  end
end

-- }}}

-- {{{ Color schemes
styles = {
  line_number = style("yellow");
  status_line = style("black,white");
  cursor = style("black,white");
  identifier = style("");
  message_info = style("");
  message_warning = style("brightyellow");
  message_error = style("brightred");
}
function style_for(name)
  return styles[name] or style("default")
end
-- }}}

-- {{{ Window
win_type_hori = "win_type_hori"
win_type_vert = "win_type_vert"
win_type_leaf = "win_type_leaf"

function win_new(typ, a, b)
  local w
  if typ == win_type_leaf then
    w = { buffer = a; scroll_line = 1 }
  elseif typ == win_type_hori then
    w = { top = a; bottom = b }
  elseif typ == win_type_vert then
    w = { left = a; right = b }
  else
    error("win_new: given invalid type")
  end

  w.typ = typ
  w.id = next_id()
  return w
end
-- }}}

-- {{{ Buffer
function buf_new(name, path)
  return {
    id = next_id();
    name = name;
    path = path;
    modified = false;
    major_mode = "normal";
    minor_modes = {};

    lines = {""};
    x = 1;
    y = 1;

    settings = {};
  }
end

function buffer_load(path)
  local path_parts = strings.split(path, "/")
  local b = buf_new(path_parts[#path_parts], path)

  ok, contents = pcall(file_read_all, path)
  if not ok then
    message(contents, "error")
    return
  end

  b.lines = strings.split(contents, "\n")
  if b.lines[#b.lines] == "" then
    table.remove(b.lines, #b.lines)
  end

  table.insert(buffers, b)
  return b
end

function buffer_save(b)
  if not b.path then
    message("Current buffer has no file path set. Use ':write <name>' to specify one", "warning")
    return
  end

  ok, err_message = pcall(file_write_all, b.path, string.join(b.lines, "\n"))
  if not ok then
    message(err_message, "error")
    return
  end

  b.modified = false
  message("Written buffer to '"..b.path.."'")
end

function buffer_set_path(b, path)
  local path_parts = strings.split(path, "/")
  b.name = path_parts[#path_parts]
  b.path = path
end

function buffer_move_to(win, x, y)
  local b = win.buffer
  b.y = math.max(1, math.min(y, #b.lines)) -- y first
  b.x = math.max(1, math.min(x, #b.lines[b.y]+1))
end

function buffer_move(win, xmov, ymov)
  local b = win.buffer
  buffer_move_to(win, win.buffer.x+xmov, win.buffer.y+ymov)
end

function buffer_move_left(win)
  buffer_move(win, -1, 0)
end
function buffer_move_right(win)
  buffer_move(win, 1, 0)
end
function buffer_move_up(win)
  buffer_move(win, 0, -1)
end
function buffer_move_down(win)
  buffer_move(win, 0, 1)
end

function buffer_move_start(win)
  buffer_move_to(win, 0, 0)
end
function buffer_move_end(win)
  buffer_move_to(win, 0, #win.buffer.lines)
end
function buffer_move_line_start(win)
  buffer_move_to(win, 0, win.buffer.y)
end
function buffer_move_line_end(win)
  local b = win.buffer
  buffer_move_to(win, #b.lines[b.y]+1, b.y)
end

-- skip word separators till 1st word char
function _buffer_move_to_word_char(win)
  local line = string.sub(win.buffer.lines[win.buffer.y], win.buffer.x)
  local index = string.find_char_index(line, word_chars)
  if index ~= 0 then
    buffer_move(win, index-1, 0)
  end
end
-- skip current word till 1st word separator
function _buffer_move_to_word_separator(win, line_extra)
  local line = string.sub(win.buffer.lines[win.buffer.y], win.buffer.x)..(line_extra or "")
  local index = string.find_char_index(line, word_sep_chars)
  if index ~= 0 then
    buffer_move(win, index-1, 0)
  end
end
function _buffer_move_word_start(win, first_call)
  local initial_x = win.buffer.x
  _buffer_move_to_word_separator(win)
  _buffer_move_to_word_char(win)
  -- if we didn't move, go to next line
  if initial_x == win.buffer.x and first_call then
    buffer_move_to(win, 0, win.buffer.y+1)
    _buffer_move_word_start(win, false)
  end
end
function buffer_move_word_start(win)
  _buffer_move_word_start(win, true)
end
function _buffer_move_word_end(win, first_call)
  local initial_x = win.buffer.x
  buffer_move(win, 1, 0) -- skip word last char
  _buffer_move_to_word_char(win)
  _buffer_move_to_word_separator(win, " ")
  if win.buffer.x > initial_x+1 then
    buffer_move(win, -1, 0)
  else
    if first_call then
      -- if we didn't move, go to next line
      buffer_move_to(win, 0, win.buffer.y+1)
      _buffer_move_word_end(win, false)
    end
  end
end
function buffer_move_word_end(win)
  _buffer_move_word_end(win, true)
end
function _buffer_move_word_start_backwards(win, first_call)
  local initial_x = win.buffer.x

  local line = string.reverse(string.sub(win.buffer.lines[win.buffer.y], 1, win.buffer.x-1))
  local index = string.find_char_index(line, word_chars)
  if index ~= 0 then
    buffer_move(win, 0-index, 0)
  end
  line = string.reverse(string.sub(win.buffer.lines[win.buffer.y], 1, win.buffer.x-1)).." "
  index = string.find_char_index(line, word_sep_chars)
  if index ~= 0 then
    buffer_move(win, 0-index+1, 0)
  end

  if initial_x == win.buffer.x and first_call then
      buffer_move_up(win)
      buffer_move_line_end(win)
      _buffer_move_word_start_backwards(win, false)
  end
end
function buffer_move_word_start_backwards(win)
  _buffer_move_word_start_backwards(win, true)
end

function buffer_move_jump_up(win)
  buffer_move(win, 0, -15)
end
function buffer_move_jump_down(win)
  buffer_move(win, 0, 15)
end

function buffer_center(win)
  win.buffer.frame_centered = true
end

function _buffer_insert(win, text)
  local b = win.buffer
  local line = b.lines[b.y]
  b.lines[b.y] = string.sub(line, 0, b.x-1)..text..string.sub(line, b.x)
  buffer_move(win, #text, 0)

  b.modified = true
end
function buffer_insert(win, k)
  local text = key_str(k)
  _buffer_insert(win, text)
end
function buffer_insert_tab(win)
  _buffer_insert(win, "    ")
end
function buffer_insert_space(win)
  _buffer_insert(win, " ")
end
function buffer_insert_return(win)
  local b = win.buffer
  local line = b.lines[b.y]
  b.lines[b.y] = string.sub(line, 0, b.x-1)
  table.insert(b.lines, win.buffer.y+1, string.sub(line, b.x))
  buffer_move_to(win, 0, b.y+1)

  b.modified = true
end
function buffer_insert_newline_up(win)
  table.insert(win.buffer.lines, win.buffer.y, "")
  buffer_move(win, 0, 0)

  win.buffer.modified = true
end
function buffer_insert_newline_down(win)
  table.insert(win.buffer.lines, win.buffer.y+1, "")
  buffer_move(win, 0, 1)

  win.buffer.modified = true
end
function buffer_delete_char(win)
  local b = win.buffer
  local line = b.lines[b.y]
  if b.x > 1 then
    b.lines[b.y] = string.sub(line, 0, b.x-1)..string.sub(line, b.x+1)
    buffer_move(win, 0, 0)
  elseif b.y > 1 then
    local prev_line_len = #b.lines[b.y-1]
    b.lines[b.y-1] = b.lines[b.y-1]..b.lines[b.y]
    table.remove(b.lines, b.y)
    buffer_move_to(win, prev_line_len+1, b.y-1)
  elseif b.x == 1 then
    b.lines[b.y] = string.sub(line, 0, b.x-1)..string.sub(line, b.x+1)
    buffer_move(win, 0, 0)
  else
    -- we are at beginning or line, first line, can't delete anything
    return
  end

  b.modified = true
end
function buffer_delete_line(win)
  table.remove(win.buffer.lines, win.buffer.y)
  buffer_move(win, 0, 0)

  win.buffer.modified = true
end
-- }}}

-- {{{ Display
function display_window_leaf(win, x, y, w, h)
  if win == current_window then
    frame(win, w, h)
  end

  local b = win.buffer
  local s = style_for("identifier")
  local sln = style_for("line_number")

  local gutter_w = #tostring(#b.lines)+1
  local screen_y = y

  -- contents
  for line = win.scroll_line, #b.lines, 1 do
    screen_write(sln, x, screen_y, pad_left(tostring(line), gutter_w-1, " "))
    screen_write(s, x+gutter_w, screen_y, b.lines[line])

    if screen_y >= h-1 then -- (-1) for status line
      break
    end

    screen_y = screen_y + 1
  end

  -- write cursor
  local sc = style_for("cursor")
  local ch = string.sub(b.lines[b.y].." ", b.x, b.x)
  screen_write(sc, x+gutter_w+b.x-1, y+b.y-win.scroll_line, ch)

  -- status line
  local ssl = style_for("status_line")
  local text = " "..b.name.." "..b.x..":"..b.y.."/"..#b.lines
  screen_write(ssl, x, y+h-1, pad_right(text, w, " "))
end

function display_window(win, x, y, w, h)
  if win.typ == win_type_leaf then
    display_window_leaf(win, x, y, w, h)
  else
    error("don't know how to display window")
  end
end

function display_bottom_bar(y, width)
  local s = style_for("message_".._message_type)
  screen_write(s, 0, y, " ".._message)
  local text_right = key_str(keys_entered)
  screen_write(s, width-#text_right-1, y, text_right)
end

function display()
  local width, height = screen_size()

  screen_clear()
  display_window(root_window, 0, 0, width, height-1)
  display_bottom_bar(height-1, width)
  screen_show()
end

-- scroll window if needed to still show cursor
function frame(win, width, height)
  local b = win.buffer

  -- After "z z" was pressed
  if b.frame_centered then
    b.frame_centered = false
    win.scroll_line = math.max(b.y - math.floor((height-1)/2), 1)
    return
  end

  -- to low
  -- (height-1) as height includes status bar
  -- (scroll_line-1) as scroll_line is 1 based
  if b.y > height-1 + win.scroll_line-1 then
    win.scroll_line = math.max(b.y - height-1 + 2, 0) + 1
  end
  -- to high
  if b.y < win.scroll_line then
    win.scroll_line = b.y
  end
end
-- }}}

-- {{{ Keymap
function keymap_new(name)
  return { id = next_id(); name = name; bindings = {}; }
end

function create_keymap(name)
  keymaps[name] = keymap_new(name)
end

function bind(keymap_name, k, func)
  if not keymaps[keymap_name] then
    message("bind: '"..keymap_name.."' is not a registered keymap", "error")
    return
  end
  if type(k) ~= "userdata" then
    message("bind: not given a key but a '"..type(k).."'", "error")
    return
  end
  keymaps[keymap_name].bindings[k] = func
end

function enter_normal_mode(win)
  if win.buffer.major_mode == "insert" then
    buffer_move_left(win)
  end
  win.buffer.major_mode = "normal"
end
function enter_insert_mode(win)
  win.buffer.major_mode = "insert"
end
function enter_command_mode(win)
  message(":")
  win.buffer.major_mode = "command"
end

catchall_key = key("c a t c h a l l")

create_keymap("normal")
bind("normal", key("h"), buffer_move_left)
bind("normal", key("j"), buffer_move_down)
bind("normal", key("k"), buffer_move_up)
bind("normal", key("l"), buffer_move_right)
bind("normal", key("g g"), buffer_move_start)
bind("normal", key("G"), buffer_move_end)
bind("normal", key("0"), buffer_move_line_start)
bind("normal", key("^"), buffer_move_line_start) -- should skip whitespace
bind("normal", key("$"), buffer_move_line_end)
bind("normal", key("w"), buffer_move_word_start)
bind("normal", key("W"), buffer_move_word_start) -- should include punctuation
bind("normal", key("e"), buffer_move_word_end)
bind("normal", key("E"), buffer_move_word_end) -- should include punctuation
bind("normal", key("b"), buffer_move_word_start_backwards)
bind("normal", key("B"), buffer_move_word_start_backwards) -- should include punctuation
bind("normal", key("C-u"), buffer_move_jump_up)
bind("normal", key("C-d"), buffer_move_jump_down)
bind("normal", key("z z"), buffer_center)
bind("normal", key("x"), buffer_delete_char)
bind("normal", key("d d"), buffer_delete_line)
bind("normal", key("i"), enter_insert_mode)
bind("normal", key(":"), enter_command_mode)
bind("normal", key("a"), chain(buffer_move_right, enter_insert_mode))
bind("normal", key("A"), chain(buffer_move_line_end, enter_insert_mode))
bind("normal", key("o"), chain(buffer_insert_newline_down, enter_insert_mode))
bind("normal", key("O"), chain(buffer_insert_newline_up, enter_insert_mode))
bind("normal", key("ESC"), function() end)
bind("normal", key("C-g"), function() end)

create_keymap("insert")
bind("insert", key("ESC"), enter_normal_mode)
bind("insert", key("RET"), buffer_insert_return)
bind("insert", key("SPC"), buffer_insert_space)
bind("insert", key("TAB"), buffer_insert_tab)
bind("insert", key("BAK2"), chain(buffer_move_left, buffer_delete_char))
bind("insert", key("DEL"), buffer_delete_char)
bind("insert", catchall_key, buffer_insert)

create_keymap("command")
function cmd_exit_command_mode(win, k)
  message("")
  enter_normal_mode(win, k)
end
bind("command", key("ESC"), cmd_exit_command_mode)
bind("command", key("C-c"), cmd_exit_command_mode)
bind("command", key("C-g"), cmd_exit_command_mode)
bind("command", key("RET"), function(win, k)
  local command_text = string.sub(message(), 2)
  local command_parts = strings.split(command_text, " ")
  local command_name = table.remove(command_parts, 1)
  local command = commands[command_name]
  if command then
    command.fn(command_parts)
    if string.sub(message(), 1, 1) == ":" then
      message("")
    end
  else
    message("Unknown command '"..command_name.."'", "warning")
  end
  enter_normal_mode(win, k)
end)
bind("command", key("BAK2"), function(win, k)
  if message() ~= ":" then
    message(string.sub(message(), 1, #message()-1))
  end
end)
bind("command", catchall_key, function(win, k)
  local text = key_str(k)
  if #text == 1 then
    message(message()..text)
  end
end)
-- }}}

-- {{{ Commands
function cmd_new(name, fn, completion_fn)
  return {
    id = next_id();
    name = name;
    fn = fn;
    completion_fn = completion_fn;
  }
end

function add_command(name, fn, completion_fn)
  commands[name] = cmd_new(name, fn, completion_fn)
end

function alt_command(alias, name)
  if not commands[name] then
    message("alias_command: no command named '"..name.."'", "error")
    return
  end
  commands[alias] = commands[name]
end


function cmd_quit()
  local b = current_window.buffer
  if b.modified then
    message("Buffer '"..b.name.."' has unsaved changes, save it first or use ':quit!'", "error")
    return
  end

  for i, v in ipairs(buffers) do
    if v.id == b.id then
      table.remove(buffers, i)
    end
  end

  if #buffers == 0 then
    -- quit when no buffer is left
    screen_quit()
    os.exit(0)
  else
    -- else, switch to buffer left
    -- TODO better window swap
    root_window = win_new(win_type_leaf, buffers[1])
    current_window = root_window
  end
end

function cmd_force_quit()
  screen_quit()
  os.exit(0)
end

function cmd_edit(args)
  local new_buffer = buffer_load(args[1])
  root_window = win_new(win_type_leaf, new_buffer)
  current_window = root_window
end

function cmd_write(args)
  local b = current_window.buffer
  if args[1] then
    buffer_set_path(b, args[1])
  end
  buffer_save(b)
end

add_command("quit", cmd_quit)
alt_command("q", "quit")
alt_command("close", "quit")
add_command("quit!", cmd_force_quit)
alt_command("q!", "quit!")
add_command("edit", cmd_edit)
alt_command("e", "edit")
alt_command("ed", "edit")
alt_command("open", "edit")
alt_command("o", "edit")
add_command("write", cmd_write)
alt_command("w", "write")
add_command("wq", chain(cmd_write, cmd_quit))
-- }}}

-- {{{ Main loop
function main()
  -- Load files in ARGS
  for i = 2, #ARGS, 1 do
    buffer_load(ARGS[i])
  end
  if #buffers == 0 then
    local scratch_buffer = buf_new("*scratch*")
    table.insert(buffers, scratch_buffer)
  end

  -- Show first buffer in root window
  root_window = win_new(win_type_leaf, buffers[1])
  current_window = root_window
  keys_entered = key("")

  local next_key = screen_next_key()
  while true do
    if next_key then
      keys_entered = key_append(keys_entered, next_key)

      local current_keymap = keymaps[current_window.buffer.major_mode]
      local catchall_handler
      local key_used = false
      for k,v in pairs(current_keymap.bindings) do
        if key_matches_part(keys_entered, k) then
          v(current_window, k)
          keys_entered = key("")
          key_used = true
          break
        end
        if key_matches(k, catchall_key) then
          catchall_handler = v
        end
      end

      if not key_used and catchall_handler then
        catchall_handler(current_window, keys_entered)
        keys_entered = key("")
      end

    end

    display()
    next_key = screen_next_key()
  end
end

main()
-- }}}
`
