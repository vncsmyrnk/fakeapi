---@class FakeApi
local M = {}

local defaults = {
  port = 8080,
}

---@class AssertOpts
---@field count integer
---@field port integer

---@param method string
---@param path string
---@param opts AssertOpts
---@return boolean
function M.assert(method, path, opts)
  opts = opts or {}
  local cmd = string.format(
    [[fakeapi assert %s %s -c %s -p %s --quiet]],
    method,
    path,
    opts.count or 1,
    opts.port or defaults.port
  )
  return os.execute(cmd) == 0
end

---@param method string
---@param path string
---@param status_code integer
---@param response_json string
---@return boolean
function M.set_endpoint(method, path, status_code, response_json)
  local cmd = string.format([[fakeapi set endpoint %s '%s' -s %s -r '%s']], method, path, status_code, response_json)
  return os.execute(cmd) == 0
end

---@class ClearRequestsOpts
---@field port integer

---@param opts ClearRequestsOpts
---@return boolean
function M.clear_requests(opts)
  opts = opts or {}
  local cmd = string.format([[fakeapi clear requests -p %s --quiet]], opts.port or defaults.port)
  return os.execute(cmd) == 0
end

return M
