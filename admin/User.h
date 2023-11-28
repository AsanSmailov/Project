#include <httplib.h> //библиотека для http
#include <xlnt/xlnt.hpp> //билблиотека для парсинга
#include <jwt-cpp/jwt.h> //подключил библиотеку на jwt токен
#include <nlohmann/json.hpp> //библиотека json
#include <cookie.h>
#include <string>
#include <sstream>
#include <cstdlib>
#include <map>
#include <ctime>
#include <iostream>
#include <vector>
#include <chrono>
#include <stack>
using namespace httplib;
using namespace std;
using json = nlohmann::json;

//Объект(имя и группа пользователя
struct object {
	string full_name;
	string group;
};

//структура пользователя
struct User {
	int github_id;
	int tg_id;
	string role;
	object user;
};

//github : token
map <int, string>open_session;

map <string, string> session;
//token : jwt, jwtsecret
map <string, map<string, string>> sessiondata;
string secret;

//функция для расшифровки jwt токена
User DecodeJWT(const string token, const string secret) {
    User user;
    user.github_id = 0;
    // Создаем объект verifier(проверка), используя алгоритм HS256
    auto verifier = jwt::verify()
        .allow_algorithm(jwt::algorithm::hs256{ secret });

    // Разбор и проверка JWT-токена. Мы передаем token и verifier в функцию decode.
    auto decoded_token = jwt::decode(token);
    verifier.verify(decoded_token);
    //парсинг времени, проверка жизни токена
    string time;
    //парсинг времени системного
    auto now = chrono::system_clock::now();
    time_t curtime = std::chrono::system_clock::to_time_t(now);


    //парсинг времени авторизации
    auto timeExp = decoded_token.get_payload_claim("expires_at").as_int();
    cout << "timeEXP: " << timeExp << "\n";
    cout << "curtime: " << curtime << "\n";
    //определение жизни токена
    if (timeExp < curtime) {
        cout << "error";
        return user;
    }
    else {
        // Извлечение полей из JWT-токена
        user.github_id = decoded_token.get_payload_claim("githubID").as_int();
        user.tg_id = decoded_token.get_payload_claim("tgID").as_int();
        user.user.full_name = decoded_token.get_payload_claim("full_name").as_string();
        user.user.group = decoded_token.get_payload_claim("group").as_string();
        user.role = decoded_token.get_payload_claim("role").as_string();
        return user;
    }

}

//функция-обработки пост запроса на ключ
void priemsecret(const Request& req, Response& res) {
    secret = req.has_param("SECRET") ? req.get_param_value("SECRET") : 0;
    cout << "jwt secret: " << secret << "\n";
}

// Функция для создания случайной строки
string generateRandomString(int length) {
    const string chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    string randomString;
    srand(time(0));
    for (int i = 0; i < length; i++) {
        randomString += chars[rand() % chars.length()];
    }
    return randomString;
}

//функция-обработчик пост запроса на jwt
void toadmin(const Request& req, Response& res) {
    string jwt = req.has_param("jwt") ? req.get_param_value("jwt") : "0";
    string usertoken = req.has_param("usertoken") ? req.get_param_value("usertoken") : "0";
    cout << "jwt token: " << jwt << "\n";
    User user = DecodeJWT(jwt, secret);
    cout << user.github_id;
    if (user.github_id == 0) {
        cout << "Error";
        res.set_content("Error", "text/plain");
        return;
    }
    cout << user.github_id;
    if (open_session.count(user.github_id) == 0 && usertoken == "0") {
        //создание токена сессии
        string tokenstr = generateRandomString(16);
        open_session[user.github_id] = tokenstr;
        session[tokenstr] = jwt;
        sessiondata[jwt]["token"] = tokenstr;
        sessiondata[jwt]["jwt_secret"] = secret;
    }
    else if (open_session.count(user.github_id) == 0 && usertoken != "0") {
        open_session[user.github_id] = usertoken;
        session[usertoken] = jwt;
        sessiondata[jwt]["token"] = usertoken;
        sessiondata[jwt]["jwt_secret"] = secret;
    }
    //создаём свой URL
    string URL = "http://10.99.8.148:8083/admin-panel?token=" + open_session[user.github_id];
    cout << URL << "\n";
    res.set_content(URL, "text/plain");
}
void updatesession(string token) {
    auto decoded_token = jwt::decode(session[token]);
    auto payload = decoded_token.get_payload_claims();
    // Готовим данные
    string secret = generateRandomString(16);
    auto ExpiresAt = std::chrono::system_clock::now() + std::chrono::minutes(60);
    auto github_id = decoded_token.get_payload_claim("githubID").as_int();
    auto tg_id = decoded_token.get_payload_claim("tgID").as_int();
    auto full_name = decoded_token.get_payload_claim("full_name").as_string();
    auto group = decoded_token.get_payload_claim("group").as_string();
    auto role = decoded_token.get_payload_claim("role").as_string();

    auto newjwt = jwt::create()
        // Тип токена
        .set_type("JWT")

        .set_payload_claim("githubID", picojson::value(int64_t{ github_id }))
        .set_payload_claim("tgID", picojson::value(int64_t{ tg_id }))
        .set_payload_claim("full_name", jwt::claim(full_name))
        .set_payload_claim("role", jwt::claim(role))
        .set_payload_claim("group", jwt::claim(group))
        .set_payload_claim("expires_at", jwt::claim(ExpiresAt))
        // Подписываем токен ключом SECRET
        .sign(jwt::algorithm::hs256{ secret });

    //Обновляем сессию

    sessiondata.erase(session[token]);
    session[token] = newjwt;
    sessiondata[newjwt]["token"] = token;
    sessiondata[newjwt]["jwt_secret"] = secret;
}

void Updateuser(const Request& req, Response& res) {
    auto cookie = Cookie::get_cookie(req, "Cookie");
    auto jwt = cookie.value;
    User user = DecodeJWT(jwt, sessiondata[jwt]["jwt_secret"]);

    if (user.role != "admin") {
        res.status = 403;
    }
    string token = sessiondata[jwt]["token"];
    


    auto data = req.has_param("data") ? req.get_param_value("data") : "0";
    auto datatype = req.has_param("datatype") ? req.get_param_value("datatype") : "0";
    auto chatid = req.has_param("chatid") ? req.get_param_value("chatid") : "0";

    httplib::Client cli("http://localhost:8080");
    std::string requestURL = "/updateData";
    httplib::Headers headers = {
        { "Content-Type", "application/x-www-form-urlencoded" }
    };
    httplib::Params params;
    params.emplace("data", data);
    params.emplace("datatype", datatype);
    params.emplace("chatid", chatid);

    auto response = cli.Post(requestURL, headers, params);


    res.set_redirect("http://10.99.8.148:8083/admin-panel?token=" + token);
}

void deleteuser(const Request& req, Response& res) {
    auto cookie = Cookie::get_cookie(req, "Cookie");
    auto jwt = cookie.value;
    User user = DecodeJWT(jwt, sessiondata[jwt]["jwt_secret"]);

    if (user.role != "admin") {
        res.status = 403;
    }
    string token = sessiondata[jwt]["token"];

    auto gitid = req.has_param("gitid") ? req.get_param_value("gitid") : "0";

    httplib::Client cli("http://localhost:8080");
    std::string requestURL = "/DeleteUser";
    httplib::Headers headers = {
        { "Content-Type", "application/x-www-form-urlencoded" }
    };
    httplib::Params params;
    params.emplace("gitid", gitid);

    auto response = cli.Post(requestURL, headers, params);

    res.set_redirect("http://10.99.8.148:8083/admin-panel?token=" + token);
}
void Exit(const Request& req, Response& res) {
    auto token = req.has_param("token") ? req.get_param_value("token") : "0";

    sessiondata.erase(session[token]);
    session.erase(token);
    cout << "test1";
    for (auto& [gitid, sessiontoken] : open_session) {
        if (sessiontoken == token) {
            cout << gitid;
            open_session.erase(gitid);
            break;
        }
    }
    cout << "test3";
    res.set_redirect("http://10.99.8.148:8083/admin-panel");
}
void adminpanel(const Request& req, Response& res) {
    auto token = req.has_param("token") ? req.get_param_value("token") : "0";
    if (session.count(token)) {
        updatesession(token);
        string usersjson;
        httplib::Client cli("http://localhost:8080");
        std::string requestURL = "/getAllUsers";
        httplib::Headers headers = {
            { "Content-Type", "application/json" }
        };
        httplib::Params params;
        if (auto response = cli.Post(requestURL, headers, params)) {
            if (response->status == 200) {
                // Получаем тело ответа
                usersjson = (response->body);
            }
            else {
                std::cout << "Status error: " << response->status << std::endl;
            }
        }
        else {
            auto err = response.error();
            std::cout << "HTTP1 error: " << httplib::to_string(err) << std::endl;
        }

        json users = json::parse(usersjson);
        string html_response;
        html_response += u8"<html>";
        html_response += u8"<header><meta charset='utf-8'></header>";
        html_response += u8"<body><h2>Загрузка нового расписания</h2>";
        html_response += u8"<iframe name='dummyframe' id='dummyframe' style='display: none;'></iframe>";
        html_response += u8"<form action = 'http://10.99.8.148:8083/upload' method = 'post' enctype = 'multipart/form-data' target='dummyframe'>";
        html_response += u8"<label for = 'file'>Выберите файл : </label>";
        html_response += u8"<input type = 'file' id = 'file' name = 'file' accept = '.xlsx'>";
        html_response += u8"<br><br>";
        html_response += u8"<input type = 'submit'></form>";

        html_response += u8"<h2>Пользователи</h2>";
        html_response += u8"<table><tr><td>Имя</td><td>Группа</td><td>Роль</td></tr>";


        for (const auto& user : users) {
            string name = user["about"]["full_name"];
            string group = user["about"]["group"];
            html_response += u8"<tr>";

            html_response += u8"<td><form action='http://10.99.8.148:8083/updateuser' method='post'>";
            html_response += u8"<input type='text' , name='data' value='";
            html_response += name;
            html_response += u8"'/>";
            html_response += u8"<input type='hidden' , name='chatid' value='" + to_string(user["tg_id"]);
            html_response += u8"'/>";
            html_response += u8"<input type='hidden' , name='datatype' value='full_name'/>";
            html_response += u8"<p><button type = 'submit'>Изменить</button></p>";
            html_response += u8"</form></td>";

            html_response += u8"<td><form action='http://10.99.8.148:8083/updateuser' method='post'>";
            html_response += u8"<input type='text' , name='data' value='";
            html_response += group;
            html_response += u8"'/>";
            html_response += u8"<input type='hidden' , name='chatid' value='" + to_string(user["tg_id"]);
            html_response += u8"'/>";
            html_response += u8"<input type='hidden' , name='datatype' value='group'/>";
            html_response += u8"<p><button type = 'submit'>Изменить</button></p>";
            html_response += u8"</form></td>";

            html_response += u8"<td><form action='http://10.99.8.148:8083/updateuser' method='post'>";
            html_response += u8"<select id='data' name = 'data'>";
            html_response += u8"<option value='student'>Студент</option>";
            html_response += u8"<option value='teacher'>Преподаватель</option>";
            html_response += u8"<option value='admin'>Администратор</option>";
            html_response += u8"</select>";
            html_response += u8"<input type='hidden' , name='chatid' value='" + to_string(user["tg_id"]);
            html_response += u8"'/>";
            html_response += u8"<input type='hidden' , name='datatype' value='role'/>";
            html_response += u8"<p><button type = 'submit'>Изменить</button></p>";
            html_response += u8"</form></td>";

            html_response += u8"<td><form action='http://10.99.8.148:8083/deleteuser' method='post'>";
            html_response += u8"<p><button type='submit' name = 'gitid' value = '" + to_string(user["github_id"]) + u8"'>Удалить</button></p>";
            html_response += u8"</form></td>";

            html_response += u8"</tr>";

        }
        html_response += u8"</table>";
        html_response += u8"<form action='http://10.99.8.148:8083/exit' method='post'>";
        html_response += u8"<input type='hidden' , name='token' value='" + token + u8"'/>";
        html_response += u8"<p><button type = 'submit'>Выйти</button></p>";
        html_response += u8"</form></body></html>";



        Cookie cookie;
        cookie.name = "Cookie";
        cookie.value = session[token];
        cookie.path = "/";
        cookie.maxAge = 3600;
        cookie.httpOnly = true;
        cookie.secure = true;
        cookie.sameSite = Cookie::SameSiteLaxMode;
        Cookie::set_cookie(res, cookie);

        res.set_content(html_response, "text/html");
    }
    else {
        auto newtoken = generateRandomString(4);
        session[newtoken] = "";
        string html_response;
        html_response += u8"<html>";
        html_response += u8"<header><meta charset='utf-8'></header>";
        html_response += u8"<body><h3>Текущая сессия не найдена.</h3>";
        html_response += u8"<p>Для продолжения работы авторизуйтесь через telegram:</p>";
        html_response += u8"<h3>https://t.me/SimpleExampeBot</h3>";
        html_response += u8"<p>Затем нажмите 'toadmin(token)' и введите следующий код:</p>";
        html_response += u8"<h3>" + newtoken + u8"</h3>";
        html_response += u8"</body></html>";
        res.set_content(html_response, "text/html");
    }
}

string convert_time(double n) {
    double totalseconds;
    int hours, minutes;

    totalseconds = n * (24 * 60 * 60);
    hours = totalseconds / 3600;
    minutes = (int(totalseconds) % 3600) / 60;

    string time = to_string(hours) + ":" + to_string(minutes);
    if (time == "7:59") return "8:00";
    else if (time == "9:49") return "9:50";
    else if (time == "13:19") return "13:20";
    else if (time == "15:0") return "15:00";

    return time;
}


void excel_parser(size_t i, const httplib::MultipartFormData& file) {
    
    std::vector<unsigned char> file_content(file.content.begin(), file.content.end());
    std::istringstream stream(std::string(file_content.begin(), file_content.end()));
    xlnt::workbook wb;
    wb.load(stream);
    
    auto ws = wb.sheet_by_index(i);
    auto rows_count = ws.calculate_dimension().height();  // Количество строк
    auto cols_count = ws.calculate_dimension().width(); // Количество столбцов
    
    json rasp;
    stack<string>week;
    stack<string>day;
    for (size_t row = 1; row <= ws.calculate_dimension().height(); row++) {      // Цикл по строкам
        for (size_t col = 1; col <= 1; col++) {   // Цикл по столбцам
            auto cur_cell = ws.cell(col, row).to_string();
            if (cur_cell == "Week:") {
                if (!week.empty()) {
                    week.pop();
                }

                week.push(ws.cell(col + 1, row).to_string());
            }
            if (cur_cell == "Day:") {
                if (!day.empty()) {
                    day.pop();
                }
                day.push(ws.cell(col + 1, row).to_string());
            }
            if (cur_cell == "Subgroup:" && ws.cell(col + 1, row).to_string() == "1") {
                auto lesson_count = ws.cell(col + 1, row + 1).to_string();

                string name, teacher, type, classroom;
                double time;
                int number;

                json cur_day;
                cur_day["week"] = week.top();
                cur_day["day"] = day.top();
                cur_day["count_of_lessons"] = atoi(lesson_count.c_str());
                cur_day["subgroup"] = atoi(ws.cell(col + 1, row).to_string().c_str());
                string lesson;
                for (size_t lesson_row = row + 3; lesson_row <= row + 36; lesson_row++) {
                    if (ws.cell(col, lesson_row).to_string() == "Number:") {
                        number = ws.cell(col + 1, lesson_row).value<int>();
                        lesson = "lesson" + to_string(number);
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Name:") {
                        name = ws.cell(col + 1, lesson_row).to_string();
                        cur_day[lesson]["name"] = name;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Teacher:") {
                        teacher = ws.cell(col + 1, lesson_row).to_string();
                        cur_day[lesson]["teacher"] = teacher;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Type:") {
                        type = ws.cell(col + 1, lesson_row).to_string();
                        cur_day[lesson]["type"] = type;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Classroom:") {
                        classroom = ws.cell(col + 1, lesson_row).to_string();
                        cur_day[lesson]["classroom"] = classroom;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Time:") {
                        time = ws.cell(col + 1, lesson_row).value<double>();
                        
                        cur_day[lesson]["time"] = convert_time(time);
                    }
                }
                rasp.push_back(cur_day);
            }
            if (cur_cell == "Subgroup:" && ws.cell(col + 2, row).to_string() == "2") {
                auto lesson_count = ws.cell(col + 2, row + 1).to_string();

                string name, teacher, type, classroom;
                double time;
                int number;

                json cur_day;
                cur_day["week"] = week.top();
                cur_day["day"] = day.top();
                cur_day["count_of_lessons"] = atoi(lesson_count.c_str());
                cur_day["subgroup"] = atoi(ws.cell(col + 2, row).to_string().c_str());
                string lesson;
                for (size_t lesson_row = row + 3; lesson_row <= row + 36; lesson_row++) {
                    if (ws.cell(col, lesson_row).to_string() == "Number:") {
                        number = ws.cell(col + 2, lesson_row).value<int>();
                        lesson = "lesson" + to_string(number);
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Name:") {
                        name = ws.cell(col + 2, lesson_row).to_string();
                        cur_day[lesson]["name"] = name;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Teacher:") {
                        teacher = ws.cell(col + 2, lesson_row).to_string();
                        cur_day[lesson]["teacher"] = teacher;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Type:") {
                        type = ws.cell(col + 2, lesson_row).to_string();
                        cur_day[lesson]["type"] = type;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Classroom:") {
                        classroom = ws.cell(col + 2, lesson_row).to_string();
                        cur_day[lesson]["classroom"] = classroom;
                    }
                    else if (ws.cell(col, lesson_row).to_string() == "Time:") {
                        time = ws.cell(col + 2, lesson_row).value<double>();
                        cur_day[lesson]["time"] = convert_time(time);
                    }
                }
                rasp.push_back(cur_day);
            }
        }
    }
    string group;
    switch (i) {
    case 0: group = "231";
        break;
    case 1: group = "232";
        break;
    case 2: group = "233";
        break;
    }
    cout << "test4";

    httplib::Client cli("http://localhost:8082");
    std::string requestURL = "/upload_schedule";

    httplib::Params params;
    auto str = rasp.dump();
    params.emplace("json", str);
    params.emplace("group", group);

    httplib::Headers headers = {
        { "Content-Type", "application/x-www-form-urlencoded" }
    };
    cout << "test5";
    auto response = cli.Post(requestURL, headers, params);
    cout << "test6";
}
//Функция приёма файла расписания и последующего его парсинга
void obrabotchik(const Request& req, Response& res) {
    auto cookie = Cookie::get_cookie(req, "Cookie");
    auto jwt = cookie.value;
    User user = DecodeJWT(jwt, sessiondata[jwt]["jwt_secret"]);

    if (user.role != "admin") {
        res.status = 403;
    }
    string token = sessiondata[jwt]["token"];
    updatesession(token);
    // Получение данных файла
    auto size = req.files.size();
    auto ret = req.has_file("file");
    if (ret) {
        const auto& file = req.get_file_value("file");
        printf("filename: %s\n", file.name.c_str());
    }
    const auto& file = req.get_file_value("file");
    
    excel_parser(0, file);
    excel_parser(1, file);
    excel_parser(2, file);

   

    Cookie new_cookie;
    new_cookie.name = "Cookie";
    new_cookie.value = session[token];
    new_cookie.path = "/";
    new_cookie.maxAge = 3600;
    new_cookie.httpOnly = true;
    new_cookie.secure = true;
    new_cookie.sameSite = Cookie::SameSiteLaxMode;
    Cookie::set_cookie(res, new_cookie);
    res.set_content("Все норм", "text/plain");
}

