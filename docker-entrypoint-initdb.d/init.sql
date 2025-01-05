CREATE TABLE rule (
  top TEXT PRIMARY KEY,
  mod TEXT
);

INSERT INTO rule VALUES 
  ('test1', 'ex'), 
  ('test2', 'rw'), 
  ('test3', 'ro'),
  ('test4', 'ex');
