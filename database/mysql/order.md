# Order(순서)

- [Order(순서)](#order순서)
  - [`ORDER BY`절 없는 경우, 기본적인 순서가 있는지?](#order-by절-없는-경우-기본적인-순서가-있는지)
    - [What is the default order of records for a SELECT statement in MySQL?](#what-is-the-default-order-of-records-for-a-select-statement-in-mysql)
    - [MySQL not using PRIMARY KEY to sort by default](#mysql-not-using-primary-key-to-sort-by-default)
    - [InnoDB B+ tree index - duplicate values](#innodb-b-tree-index---duplicate-values)

## `ORDER BY`절 없는 경우, 기본적인 순서가 있는지?

### [What is the default order of records for a SELECT statement in MySQL?](https://dba.stackexchange.com/a/6053)

Reposting [my answer](https://dba.stackexchange.com/questions/5774/why-is-ssms-inserting-new-rows-at-the-top-of-a-table-not-the-bottom/5775#5775) to a similar question regarding SQL Server:

In the SQL world, order is not an inherent property of a set of data. Thus, you get no guarantees from your RDBMS that your data will come back in a certain order -- or even in a consistent order -- unless you query your data with an ORDER BY clause.

So, to answer your question:

- MySQL sorts the records however it wants without any guarantee of consistency.
- If you intend to rely on this order for anything, you must specify your desired order using `ORDER BY`. To do anything else is to set yourself up for unwelcome surprises.

This is a property of all SQL, not just MySQL. The relevant text in the [SQL-92 spec](http://www.contrib.andrew.cmu.edu/~shadow/sql/sql1992.txt) is:

> If an <order by clause> is not specified, then the ordering of the rows of Q is implementation-dependent.

There are similar bits of text in the spec for cursors.

### [MySQL not using PRIMARY KEY to sort by default](https://dba.stackexchange.com/questions/179767/mysql-not-using-primary-key-to-sort-by-default/179769#179769)

### [InnoDB B+ tree index - duplicate values](https://stackoverflow.com/a/38284670)
