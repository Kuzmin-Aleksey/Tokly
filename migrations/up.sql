create table `groups`
(
    id        int auto_increment primary key,
    lap_id    int       not null,
    create_at timestamp not null
);

create table detections
(
    id         int auto_increment primary key,
    group_id   int         not null,
    image_uid  tinyblob    not null,
    class      varchar(45) not null,
    is_problem tinyint     not null,
    constraint detection_to_group
        foreign key (group_id) references `groups` (id)
            on delete cascade
);

create index detection_to_group_idx
    on detections (group_id);
