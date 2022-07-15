(assert (> (time/format (time/now) 'timestamp) 0))
(let [now (time/parse 1656988237)]
     (assert (= 1656988237 (time/format now 'timestamp)))
     (assert (= 1656988237000 (time/format now 'timestamp-ms)))
     (assert (= "2022-07-05" (time/format now "2006-01-02"))))

(let [now (time/parse "2006-01-02 15:04:05" "2001-10-01 00:00:00" "Asia/Shanghai")]
     (assert (= 1001865600 (time/format now 'timestamp)))
     (assert (= "2001-10-01" (time/format (time/parse 1001865600) "2006-01-02")))
     (assert (= "2001-10-01" (time/format (time/parse "2006-01-02 15:04:05" "2001-10-01 00:00:00") "2006-01-02")))
     )

(assert (> (time/now) (time/parse "2001-10-01 00:00:00")))
(assert (> (time/parse 1656988237) (time/parse 1656988236)))
(assert (= (time/parse 1656988237) (time/parse 1656988237)))
