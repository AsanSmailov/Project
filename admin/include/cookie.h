#ifndef COOKIE_H
#define COOKIE_H

#include <vector>
#include <string>
#include <httplib.h>

using namespace httplib;

struct Cookie {
	/// Ќазвание (ключ) cookie
	std::string name = "";
	/// «начение cookie
	std::string value = "";
	/// ”станавливает домен, дл€ которого будет доступен cookie
	std::string domain = "";
	/// ”станавливает путь на сервере, дл€ которого будет доступен cookie
	std::string path = "";
	/// > 0 - устанавливает продолжительность срока действи€ cookie в секундах; == 0 (по умолчанию) - cookie живут пока не закрыта вкладка; < 0 - удалить немедленно
	int maxAge = 0;
	/// «апрещает доступ к cookie из JavaScript, что помогает предотвратить атаки на кражу сессий
	bool httpOnly = true;
	/// ”казывает, что cookie должен быть отправлен только по защищенному (HTTPS) соединению
	bool secure = false;
	/// ”казывает, как cookie должен быть отправлен в запросах, св€занных с кросс-сайтовыми запросами
	std::string sameSite = SameSiteNoneMode;

	/// ≈сли установлено значение "Strict", cookie не будет отправл€тьс€ в кросс-сайтовых запросах. ќн будет отправл€тьс€ только в случае, если текущий сайт совпадает с сайтом, на котором установлен cookie
	static const std::string SameSiteStrictMode;
	/// ≈сли установлено значение "Lax", cookie будет отправл€тьс€ в кросс-сайтовых запросах, если это "безопасный" HTTP-метод, такой как GET, HEAD или OPTIONS, и если это действие инициировано пользователем (например, щелчок по ссылке). ќднако, если запрос инициирован из скрипта, cookie не будет отправл€тьс€
	static const std::string SameSiteLaxMode;
	/// ≈сли установлено значение "None", cookie будет отправл€тьс€ во всех кросс-сайтовых запросах. ќднако дл€ использовани€ этого значени€ требуетс€ установка атрибута "Secure", чтобы cookie мог отправл€тьс€ только по защищенному (HTTPS) соединению
	static const std::string SameSiteNoneMode;

	/// Cookie -> string
	std::string to_string() const;
	/// ‘ормирует HTTP заголовок в котором просит клиент установить cookie и добавл€ет его в ответ сервера res
	static void set_cookie(Response& res, const Cookie& cookie);
	/// ѕытаетс€ найти в запросе req куки с именем name. ≈сли таких нет, возвращаетс€ пустой объект, если name встречетс€ больше одного раза, то берЄтс€ первое которое прислал клиент
	static Cookie get_cookie(const Request& req, const std::string& name);
};

#endif // COOKIE_H