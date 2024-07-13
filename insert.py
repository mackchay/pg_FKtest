import psycopg
with psycopg.connect("host=localhost port=5432 dbname=postgres user=postgres") as conn:

  with conn.transaction() as tx:
    conn.execute("insert into t values (1)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (7)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (6)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (5)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (4)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (3)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (2)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (1)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (7)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (6)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (5)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (4)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (3)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (2)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (1)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (7)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (6)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (1)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (7)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (6)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (5)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (4)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (1)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (7)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (6)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (5)", prepare=True)
  with conn.transaction() as tx:
    conn.execute("insert into t values (4)", prepare=True)