CREATE TABLE usertbl (
  name varchar(10)
);

CREATE TABLE grouptbl (
  name varchar(10)
);

INSERT INTO usertbl values('user123');
INSERT INTO grouptbl values('user');

SELECT * FROM usertbl;
SELECT * FROM grouptbl; 
