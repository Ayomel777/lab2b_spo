def analyze(x)
    result = 0

    if x > 0
        case x
        when 1
            if x % 2 == 0
                while x > 0
                    if x > 5
                        result += x
                    else
                        result += 1
                    end
                    x -= 1
                end
            else
                result = 1
            end
        when 2
            result = x * 2
        when 3
            result = x + 3
        when 4
            result = x / 2
        end
    else
        result = -1
    end

    return result
end
