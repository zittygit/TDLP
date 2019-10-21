create table usergroup (gid int(8) unsigned not null auto_increment, groupname varchar(128) not null, primary key (gid), unique key groupname (groupname)) engine=InnoDB auto_increment=1001 default charset=utf8;
create table user (uid int(8) unsigned not null auto_increment, username varchar(128) not null, gid int(8) unsigned default null, role varchar(8) not null default "", email varchar(128) not null, createtime datetime not null, primary key (uid), unique key username (username), key gid (gid), constraint user_gid foreign key (gid) references usergroup (gid) on delete set null on update no action) engine=InnoDB auto_increment=1001 default charset=utf8;
create table app (aid int(8) unsigned not null auto_increment, appname varchar(128) not null, path varchar(128) not null, info text not null, primary key (aid), unique key appname (appname)) engine=InnoDB auto_increment=1 default charset=utf8;
create table instance (iid int(8) unsigned not null auto_increment, instancename varchar(128) not null, aid int(8) unsigned default null, uid int(8) unsigned not null, cid int(8) unsigned not null, state tinyint(1) not null default 0, createtime datetime not null, deletetime datetime default null, primary key (iid), key instancename (instancename), key aid (aid), key uid (uid), constraint instance_aid foreign key (aid) references app (aid) on delete set null on update no action, constraint instance_uid foreign key (uid) references user (uid) on delete cascade on update no action) engine=InnoDB auto_increment=1 default charset=utf8;
create table config (cid int(8) unsigned not null auto_increment, iid int(8) unsigned not null, starttime datetime not null, endtime datetime default null, param text not null, primary key (cid), key iid (iid), constraint config_iid foreign key (iid) references instance (iid) on delete cascade on update no action) engine=InnoDB auto_increment=1 default charset=utf8;
create table proxy (pid int(8) unsigned not null auto_increment, proxyname varchar(128) not null, iid int(8) unsigned not null, firstport int(8) unsigned not null, secondport int(8) unsigned not null, httpurl varchar(256) not null, websocketurl varchar(256) not null, primary key (pid), key iid (iid), constraint proxy_iid foreign key (iid) references instance (iid) on delete cascade on update no action) engine=InnoDB auto_increment=1 default charset=utf8;