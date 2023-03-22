CREATE TABLE karmine (
  uuid          VARCHAR(255) NOT NULL,
  aeskey        VARCHAR(255), 
  xorkey1       VARCHAR(255),
  xorkey2       VARCHAR(255),
  PRIMARY KEY (uuid)
);

CREATE TABLE kreds (
  id integer primary key AUTOINCREMENT,
  platform  VARCHAR(255),
  site_url  VARCHAR(255),
  uname     VARCHAR(255),
  pass      VARCHAR(255) 
); 

CREATE TABLE kmdstack (
  cmd_id integer primary key AUTOINCREMENT,
  staged_cmd    VARCHAR(1023)
);