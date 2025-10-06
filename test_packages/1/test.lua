---comment
---@param a string this does something https://pkg.go.dev/encoding/gob#Decoder
---@param b integer
---@return string
---@return integer
function _ShouldKill(a, b)
	
	print(a)

	if b > 10 then
		return "yabba dabba doo!", 0
	end

	return "xyz", 5
end