#include <iostream>
#include "User.h" //моя библиотека пользователя
#include <jwt-cpp/jwt.h> //подключил библиотеку на jwt токен
#include <httplib.h> //подключил библиотеку на http в C++
#include <nlohmann/json.hpp> //библиотека json
#include <windows.h>;

using namespace httplib;
using namespace std;




int main() {
    SetConsoleOutputCP(65001);
    //создание сервера
    Server administration;
    
    administration.Post("/deleteuser", deleteuser);

    administration.Post("/updateuser", Updateuser);

    //прием пост запроса на jwt
    administration.Post("/toadmin", priemjwt);

    //приём пост запроса на ключ
    administration.Post("/getSecret", priemsecret);

    administration.Get("/admin-panel", adminpanel);

    //обработка запроса с фалом рассписания от сервера
    administration.Post("/upload", obrabotchik);
    
    //айпи локальный сервера
    administration.listen("localhost", 8083);
    return 1;
}