-- The schema, in case you want to run the code and verify that it works
create table messages (
    id bigint generated always as identity,
    msg text
);

-- Some example data
insert into messages (msg) values ('Hello Sailor'), ('Your ad here');
