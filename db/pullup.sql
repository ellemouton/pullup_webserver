create table if not exists users (
    id int auto_increment,
    username text not null,
    password blob not null,
   
    primary key (id)
);

create table if not exists pullups (
    id int auto_increment,
    user_id int not null, 
    day date not null,
    pullups int not null,

    primary key (id),
    foreign key (user_id) references users(id) 
)