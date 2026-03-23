package = "fakeapi"
version = "0.0.1-1"
source = {
  url = "git+ssh://git@github.com/vncsmyrnk/fakeapi.git",
}
description = {
  homepage = "https://github.com/vncsmyrnk/fakeapi",
  license = "GPL-3.0",
}
build = {
  type = "builtin",
  modules = {
    fakeapi = "api/lua/fakeapi.lua",
  },
}
