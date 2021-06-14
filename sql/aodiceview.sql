create or replace view aodiceview as 
select a.id as id,
        a.name as name,
       aodice.date as date
from authors a
   inner join aodice
      on aodice.author_id = a.id;