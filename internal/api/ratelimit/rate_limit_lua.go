package ratelimit

import "github.com/go-redis/redis"

var RateLimitLua = redis.NewScript(`
	local sessionId = KEYS[1]
	local perSecondLimit = tonumber(ARGV[1])
	local minutelyLimit = tonumber(ARGV[2])
	local hourlyLimit = tonumber(ARGV[3])
	local dailyLimit = tonumber(ARGV[4])
	local monthlyLimit = tonumber(ARGV[5])

	local perSecondKey = ARGV[6]
	local minutelyKey = ARGV[7]
	local hourlyKey = ARGV[8]
	local dailyKey = ARGV[9]
	local monthlyKey = ARGV[10]

	--local isRevoked = redis.call("GET","revoked_session:" .. sessionId)
	--if isRevoked then
	--	return {false, "revoked"}
	--end

	local current

	-- seconds limit
	if perSecondLimit > 0 then
		local secondKey = sessionId .. ":" .. perSecondKey
		current = redis.call("INCR",secondKey)
		if tonumber(current) <= perSecondLimit then
			redis.call("EXPIRE",secondKey,1)
		else
			return {false, "second"}
		end
	end

	-- minute limit
	if minutelyLimit > 0 then
		local minuteKey = sessionId .. ":" .. minutelyKey
		current = redis.call("INCR",minuteKey)
		if tonumber(current) <= minutelyLimit then
			redis.call("EXPIRE",minuteKey,60)
		else
			return {false, "minute"}
		end
	end

	-- hour limit
	if hourlyLimit > 0 then
		local hourlyKey = sessionId .. ":" .. hourlyKey
		current = redis.call("INCR",hourlyKey)
		if tonumber(current) <= hourlyLimit then
			redis.call("EXPIRE",hourlyKey,60*60)
		else
			return {false, "hour"}
		end
	end

	-- daily limit
	if dailyLimit > 0 then
		local dailyKey = sessionId .. ":" .. dailyKey
		current = redis.call("INCR",dailyKey)
		if tonumber(current) <= dailyLimit then
			redis.call("EXPIRE",dailyKey,60*60*24)
		else
			return {false, "day"}
		end
	end

	-- month limit
	if monthlyLimit > 0 then
		local monthlyKey = sessionId .. ":" .. monthlyKey
		current = redis.call("INCR",monthlyKey)
		if tonumber(current) <= monthlyLimit then
			redis.call("EXPIRE",monthlyKey,60*60*24*30)
		else
			return {false, "month"}
		end
	end

	return {true, ""}
`)
