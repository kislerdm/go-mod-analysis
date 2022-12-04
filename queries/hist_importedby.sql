WITH
    d AS (
        SELECT CASE WHEN ARRAY_LENGTH(importedby) >= 50000  THEN '13. 50000+'
                    WHEN ARRAY_LENGTH(importedby) >= 20000  THEN '12. 20000+'
                    WHEN ARRAY_LENGTH(importedby) >= 10000  THEN '11. 10000+'
                    WHEN ARRAY_LENGTH(importedby) >= 5000   THEN '10. 5000+'
                    WHEN ARRAY_LENGTH(importedby) >= 1000   THEN '09. 1000+'
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
