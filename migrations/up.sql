create table `groups`
(
    id        int auto_increment
        primary key,
    lap_id    varchar(45) not null,
    create_at timestamp   not null
);

create table detections
(
    id        int auto_increment
        primary key,
    group_id  int         not null,
    image_uid tinyblob    not null,
    class     varchar(45) not null,
    constraint detection_to_group
        foreign key (group_id) references `groups` (id)
            on delete cascade
);

create table detection_rects
(
    id           int auto_increment
        primary key,
    detection_id int   not null,
    width        int   not null,
    height       int   not null,
    x0           int   not null,
    y0           int   not null,
    x1           int   not null,
    y1           int   not null,
    confidence   float not null,
    constraint rect_to_detection
        foreign key (detection_id) references detections (id)
            on delete cascade
);

create index rect_to_detection_idx
    on detection_rects (detection_id);

create index detection_to_group_idx
    on detections (group_id);

create table lap_config
(
    lap_id int           not null,
    class  varchar(45)   not null,
    value  int default 0 not null,
    primary key (lap_id, class)
);

