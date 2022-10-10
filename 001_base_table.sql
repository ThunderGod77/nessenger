-- Write your migrate up statements here
create table users (
                       userId varchar(50) PRIMARY KEY,
                       username varchar(50) UNIQUE NOT NULL,
                       password text  NOT NULL,
                       createdAt timestamptz NOT NULL
);

create table friends(
                        userId varchar(50) NOT NULL,
                        friendId varchar(50) NOT NULL,
                        status varchar(50) NOT NULL,
                        PRIMARY KEY (userId,friendId),
                        CONSTRAINT fk_user Foreign key (userId) references users(userId) on delete cascade,
                        CONSTRAINT fk_friend Foreign key (friendId) references users(userId) on delete cascade
);

create table messages(
                         messageId varchar(50) PRIMARY KEY,
                         senderId varchar(50) NOT NULL,
                         receiverId varchar(50) NOT NULL,
                         content text ,
                         sentAt timestamptz NOT NULL,
                         CONSTRAINT fk_sender foreign key (senderId) references users(userId) on delete cascade,
                         constraint fk_receiver foreign key (receiverId)	 references users(userId) on delete cascade

);

---- create above / drop below ----
drop table messages;
drop table friends;
drop table users;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
