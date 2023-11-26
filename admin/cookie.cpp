#include "cookie.h"

/*
* 
*  Функции - хелперы
* 
**/

std::string trim_string(const std::string& str) {
	size_t start = 0, last = str.size() - 1;
	while (start <= last && str[start] == ' ') start++;
	while (start <= last && str[last] == ' ') last--;

	return start <= last ? str.substr(start, last - start + 1) : "";
}

std::tuple<std::string, std::string> cut_string(const std::string& str, const std::string& sep) {
	auto n = str.find(sep);
	if (n == std::string::npos) return { str, "" };
	return { str.substr(0, n), str.substr(n + sep.size(), str.size() - (n + sep.size())) };
}

bool is_cookie_name_valid(std::string name) {
	if (name == "") return false;
	return true;
}

bool valid_cookie_value_char(unsigned char b) {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\';
}

// Strip the quotes, if present.
std::string parse_cookie_value(std::string value, bool allowDoubleQuote) {
	if (allowDoubleQuote && value.size() > 2 && value[0] == '"' && value[value.size() - 1] == '"') {
		value = value.substr(1, (value.size() - 1) - 1);
	}
	for (int i = 0; i < value.size(); i++) {
		if (not valid_cookie_value_char(value[i])) {
			throw "valid_cookie_value_char error";
			return "";
		}
	}
	return value;
}

std::vector<Cookie> read_cookies(const Request& req, const std::string filter) {
	std::vector<Cookie> cookies;

	auto count = req.get_header_value_count("Cookie");
	if (count == 0) return cookies;

	std::vector<std::string> lines(count);
	for (size_t i = 0; i < count; i++) lines[i] = req.get_header_value("Cookie", i);

	for (std::string& line : lines) {
		line = trim_string(line);

		std::string part;
		while (line.size() > 0) {
			std::tie(part, line) = cut_string(line, ";");
			if (part == "") continue;

			auto [name, value] = cut_string(part, "=");
			if (not is_cookie_name_valid(name)) continue;

			if (filter != "" && filter != name) continue;

			value = parse_cookie_value(value, true);

			Cookie cookie;
			cookie.name = name;
			cookie.value = value;
			cookies.push_back(cookie);
		}
	}

	return cookies;
}

/*
*
*  Методы класса Cookie
*
**/

std::string Cookie::to_string() const {
	if (name == "") return "";
	return name + "=" + value +
		(domain == "" ? "" : "; Domain=" + domain) +
		(path == "" ? "" : "; Path=" + path) +
		(maxAge == 0 ? "" : "; Max-Age=" + std::to_string(maxAge)) +
		(not httpOnly ? "" : "; HttpOnly") +
		(not secure ? "" : "; Secure") +
		(sameSite == SameSiteNoneMode ? "" : "; SameSite=" + sameSite);
}

void Cookie::set_cookie(Response& res, const Cookie& cookie) {
	if (auto v = cookie.to_string(); v != "") {
		res.set_header("Set-Cookie", v);
	}
}

Cookie Cookie::get_cookie(const Request& req, const std::string& name) {
	if (name == "") {
		return Cookie();
	}
	auto list = read_cookies(req, name);
	for (auto& cookie : list) {
		return cookie;
	}
	return Cookie();
}

/*
*
*  Статические константы класса Cookie
*
**/

const std::string Cookie::SameSiteStrictMode = "Strict";
const std::string Cookie::SameSiteLaxMode = "Lax";
const std::string Cookie::SameSiteNoneMode = "None";