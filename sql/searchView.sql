create or replace view searchview as 
select authors.id as authorid,
       authors.name,
       quotes.id as quoteid,
       quotes.quote as quote,
       quotes.isicelandic as isicelandic
from authors
   inner join quotes
      on authors.id = quotes.authorid;