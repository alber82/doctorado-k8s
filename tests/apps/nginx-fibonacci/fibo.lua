local args = ngx.req.get_uri_args()
local n = tonumber(args["n"]) or 0

local function fib(x)
    if x < 2 then return x end
    local a, b = 0, 1
    for i = 2, x do
        a, b = b, a + b
    end
    return b
end

ngx.header.content_type = 'application/json'
ngx.say(string.format('{"n":%d,"result":%d}', n, fib(n)))