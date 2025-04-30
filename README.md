# Bet system

# Uml diagram
![Скриншот](golang-uml.jpg)

# Erd diagram
![Скриншот](golang-erd.jpg)


Team: Arlan Zhumagulov

Sanzhar Ashirbekov

Aral Askarov

![alt text](image.png)


aralcracker@WS-13815:~/golang$ curl -X POST http://golang.medhelper.xyz/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "arak@gmail.com",
    "username": "arak",
    "password": "123"
  }'
{"message":"User registered successfully"}






aralcracker@WS-13815:~/golang$ curl -X POST http://golang.medhelper.xyz/login   -H "Content-Type: application/json"   -d '{
    "email": "arak@gmail.com",
    "password": "123"
  }'
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYyMTY5MzQsInVzZXJfdXVpZCI6IjQ0MDExMTg2LTFjNmEtNDYxZi1hYWI0LTk5MzZkYzUyODhjNyJ9.By3Qh6viODLu15OhV8WEn_rrtBZ6_kI3a3yFD_Ur_jE"}

curl -X POST http://golang.medhelper.xyz/dep/withdrawal   -d "access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYyNjU4NDAsInVzZXJfdXVpZCI6Ijk3MDRiNjg5LTRlYjktNDI0YS1iMDc2LWQ0MWI2ZmE0MWY4YiJ9.lNy58lk-r8BWxdKiBOyEDhY1XLM-wXW0AfJhhvIULqI"   -d "card_number=8996450932285594"  -d "money=1"
{"message":"Balance successfully replenished to card back"}


curl -X POST http://golang.medhelper.xyz/dep/balance   -d "access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYyNjU4NDAsInVzZXJfdXVpZCI6Ijk3MDRiNjg5LTRlYjktNDI0YS1iMDc2LWQ0MWI2ZmE0MWY4YiJ9.lNy58lk-r8BWxdKiBOyEDhY1XLM-wXW0AfJhhvIULqI"   -d "card_number=8996450932285594"   -d "card_owner=Arak Askaris"   -d "cvv=141"   -d "money=125"

curl -X GET http://golang.medhelper.xyz/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYyNjk3OTYsInVzZXJfdXVpZCI6Ijk3MDRiNjg5LTRlYjktNDI0YS1iMDc2LWQ0MWI2ZmE0MWY4YiJ9.JjHV3QttozEA74KdgZTkcR0g_0_ALNJFE2F9bowqxMY" \
  -H "Content-Type: application/json"
{"uuid":"9704b689-4eb9-424a-b076-d41b6fa41f8b","username":"arak","email":"arak@gmail.com"}


curl http://localhost:8080/api/v1/events

MYSQL
sudo kubectl exec -it mysql-0 -- bash
 mysql -u root -p
USE mydatabase
SHOW tables;
SELECT * FROM users;

postgres
sudo kubectl exec -it postgres-0 -- bash
psql -U user -d golangdb
SELECT * FROM users;


curl -X GET http://golang.medhelper.xyz/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDYyODAwNzIsInVzZXJfdXVpZCI6Ijg2NWUwYTA3LTUwMDctNGIyYy1hZTc4LTY1Njk4MTRmMzU0YSJ9.wDe6jy7zdmgIgxNM9OtujnvUzGr4LLXYqUWoAp-Ag3U" \
  -H "Content-Type: application/json"