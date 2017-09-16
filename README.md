<<<<<<< HEAD
# ArticleDB

Simple web-page that shows articles written by users.
=======
# structapp

Simple web-page showing articles written by users.
>>>>>>> 7476a9d9ff3421d25f312449c6fb706706b69ce8

Bootstrap and JQuery come from their respective official CDNs.

## Installing

### MySQL

```mysql
create table Users
(
  Email varchar(255) not null,
  Username varchar(32) not null,
  Password binary(60) not null,
  UserURL varchar(10) not null
    primary key,
  CreatedOn timestamp default CURRENT_TIMESTAMP not null
)
;

create index Users_UserURL_index
  on Users (UserURL)
;


create table Posts
(
  Title varchar(255) not null,
  Content text not null,
  PostURL varchar(10) not null
    primary key,
  CreatorURL varchar(10) not null,
  CreatedOn timestamp default CURRENT_TIMESTAMP not null,
  constraint Posts_Users_UserURL_fk
  foreign key (CreatorURL) references Users (UserURL)
    on update cascade on delete cascade
)
;

create index Posts_PostURL_index
  on Posts (PostURL)
;

create index Posts_Users_UserURL_fk
  on Posts (CreatorURL)
;
```
