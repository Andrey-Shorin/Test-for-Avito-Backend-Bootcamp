create type user_type as enum ('client', 'moderator');
create type request_state as enum ('created', 'approved', 'declined','on moderation');

create table  houses(id SERIAL  primary key , address text , year INT,
    developer text, created_at timestamp,update_at timestamp);

create table  flats(flatId  SERIAL primary key , houseId INTEGER REFERENCES houses (id) ON DELETE CASCADE,
    price INT CHECK(price >= 0), rooms INT CHECK(price >= 0),status request_state);

create table users(userID text UNIQUE, email text UNIQUE, password text, type user_type);

create table tokens(id SERIAL  primary key ,userID text  REFERENCES users (userID) ON DELETE CASCADE,
    token text UNIQUE,type user_type, created_at timestamp);

create table subscribe( email text 
,  houseId INTEGER REFERENCES houses (id) ON DELETE CASCADE);

create index ix_users on users (userID);
create index ix_houses on houses (id);
create index ix_flats_houseId on flats (houseId);
create index ix_flats_flatId on flats (flatId);
create index ix_token on tokens (token);
create index ix_tokenID on tokens (userID);
create index ix_subscribe on subscribe (houseId);