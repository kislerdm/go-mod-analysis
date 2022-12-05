WITH
    d AS (
        SELECT CASE WHEN ARRAY_LENGTH(importedby) >= 50000  THEN '26. 50000+'
                    WHEN ARRAY_LENGTH(importedby) >= 40000  THEN '25. 40000+'
                    WHEN ARRAY_LENGTH(importedby) >= 30000  THEN '24. 30000+'
                    WHEN ARRAY_LENGTH(importedby) >= 20000  THEN '23. 20k-30k'
                    WHEN ARRAY_LENGTH(importedby) >= 10000  THEN '22. 10k-20k'
                    WHEN ARRAY_LENGTH(importedby) >= 9000   THEN '21. 9000+'
                    WHEN ARRAY_LENGTH(importedby) >= 8000   THEN '20. 8000+'
                    WHEN ARRAY_LENGTH(importedby) >= 7000   THEN '19. 7000+'
                    WHEN ARRAY_LENGTH(importedby) >= 6000   THEN '18. 6000+'
                    WHEN ARRAY_LENGTH(importedby) >= 5000   THEN '17. 5000+'
                    WHEN ARRAY_LENGTH(importedby) >= 4000   THEN '16. 4000+'
                    WHEN ARRAY_LENGTH(importedby) >= 3000   THEN '15. 3000+'
                    WHEN ARRAY_LENGTH(importedby) >= 2000   THEN '14. 2000+'
                    WHEN ARRAY_LENGTH(importedby) >= 1000   THEN '13. 1000+'
                    WHEN ARRAY_LENGTH(importedby) >= 900    THEN '12. 900+'
                    WHEN ARRAY_LENGTH(importedby) >= 800    THEN '11. 800+'
                    WHEN ARRAY_LENGTH(importedby) >= 700    THEN '10. 700+'
                    WHEN ARRAY_LENGTH(importedby) >= 600    THEN '09. 600+'
                    WHEN ARRAY_LENGTH(importedby) >= 500    THEN '08. 500+'
                    WHEN ARRAY_LENGTH(importedby) >= 400    THEN '07. 400+'
                    WHEN ARRAY_LENGTH(importedby) >= 300    THEN '06. 300+'
                    WHEN ARRAY_LENGTH(importedby) >= 200    THEN '05. 200+'
                    WHEN ARRAY_LENGTH(importedby) >= 100    THEN '04. 100+'
                    WHEN ARRAY_LENGTH(importedby) >= 50     THEN '03. 50+'
                    WHEN ARRAY_LENGTH(importedby) >= 10     THEN '02. 10+'
                    WHEN ARRAY_LENGTH(importedby) >= 1      THEN '01. 1+'
                                                            ELSE '00. 0'
               END                  AS group_importedby
             , COUNT(DISTINCT path) AS cnt_module
          FROM `go-mod-analysis.raw.pkggodev`
          GROUP BY 1
          ORDER BY 1
    )

SELECT SPLIT(group_importedby, '. ')[OFFSET(1)] AS group_importedby
     , cnt_module
FROM d
;
