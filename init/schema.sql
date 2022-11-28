drop table if exists events;
drop table if exists calendar;
drop table if exists user;

create table if not exists calendar (
    uuid varchar(255) primary key,
    user_token varchar(255) not null,
    date integer not null unique,
    foreign key(user_token) references user(token)
);

create table if not exists events (
    uuid varchar(255) primary key,
    calendar_uuid varchar(255) not null,
    title varchar(255) not null,
    start_time integer not null,
    end_time integer not null,
    foreign key(calendar_uuid) references calendar(uuid)
);

create table if not exists user (
    token varchar(255) primary key
);
