---@class FakeApi
local M = {}

---@param port integer
---@param config_filepath string
function M.start_server_docker(port, config_filepath)
  local cmd = string.format(
    [[docker run --rm -d \
--name fakeapi-server \
-v %s:/data/config.json \
-p 8080:%s \
vncsmyrnk/fakeapi \
/data/config.json >/dev/null 2>&1]],
    config_filepath,
    port
  )
  return os.execute(cmd) == 0
end

function M.stop_server_docker()
  local cmd = [[docker stop fakeapi-server >/dev/null 2>&1]]
  return os.execute(cmd) == 0
end

---@class AssertOpts
---@field count integer
---@field port integer

---@param method string
---@param uri string
---@param opts AssertOpts
---@return boolean
function M.assert(method, uri, opts)
  local cmd = string.format([[fakeapi assert %s %s -c %s --quiet]], method, uri, opts.count or 1)
  return os.execute(cmd) == 0
end

return M
