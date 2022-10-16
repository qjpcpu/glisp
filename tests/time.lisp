(assert (> (time/format (time/now) 'timestamp) 0))
(let [now (time/parse 1656988237)]
     (assert (= 1656988237 (time/format now 'timestamp)))
     (assert (= 1656988237000 (time/format now 'timestamp-ms)))
     (assert (= "2022-07-05" (time/format now "2006-01-02"))))

(let [now (time/parse "2006-01-02 15:04:05" "2001-10-01 00:00:00" "Asia/Shanghai")]
     (assert (= 1001865600 (time/format now 'timestamp)))
     (assert (= "2001-10-01" (time/format (time/parse 1001865600) "2006-01-02" "Asia/Shanghai")))
     (assert (= "2001-10-01" (time/format (time/parse "2006-01-02 15:04:05" "2001-10-01 00:00:00") "2006-01-02")))
     )

(assert (> (time/now) (time/parse "2001-10-01 00:00:00")))
(assert (> (time/parse 1656988237) (time/parse 1656988236)))
(assert (= (time/parse 1656988237) (time/parse 1656988237)))

(let [tm (time/parse 1656988237)]
     (assert (= 2022 (time/year tm)))
     (assert (= 7 (time/month tm)))
     (assert (= 5 (time/day tm)))
     (assert (= 10 (time/hour tm)))
     (assert (= 30 (time/minute tm)))
     (assert (= 37 (time/second tm)))
     (assert (= 2 (time/weekday tm)))
     (let [tm2 (time/add-date tm 1 1 1)]
       (assert (= 2023 (time/year tm2)))
       (assert (= 8 (time/month tm2)))
       (assert (= 6 (time/day tm2)))
     )
     (let [tm2 (time/add tm 1 'year)]
       (assert (= 2023 (time/year tm2))))
     (let [tm2 (time/add tm 1 'month)]
       (assert (= 8 (time/month tm2))))
     (let [tm2 (time/add tm 1 'day)]
       (assert (= 6 (time/day tm2))))
     (let [tm2 (time/add tm 3 'hour)]
       (assert (= 13 (time/hour tm2))))
     (let [tm2 (time/add tm 3 'minute)]
       (assert (= 33 (time/minute tm2))))
     (let [tm2 (time/add tm 3 'second)]
       (assert (= 40 (time/second tm2))))
)

(let [t1 (time/parse 1656988237) t2 (time/parse 1656988230)]
     (assert (= 7 (time/sub t1 t2)))
)


(def t2 (time/now))
(assert (= (time/format t2 'timestamp) (time/format (time/parse (time/format t2 'timestamp) 'timestamp) 'timestamp)))
(assert (= (time/format t2 'timestamp-micro) (time/format (time/parse (time/format t2 'timestamp-micro) 'timestamp-micro) 'timestamp-micro)))
(assert (= (time/format t2 'timestamp-nano) (time/format (time/parse (time/format t2 'timestamp-nano) 'timestamp-nano) 'timestamp-nano)))


;; parse ok
(time/parse "2022-09-19T08:25:26Z")
(time/parse "2006-01-02T15:04:05+07:00")
(time/parse "2001-01-01")
(time/parse "01:01:01")

(let [now (time/now)]
     (assert (!=
                (time/format now "15:04:05" "Asia/Shanghai")
                (time/format now "15:04:05" "UTC")
              ))
     )

(assert (not (zero? (time/now))))
(assert (zero? (time/parse -62135596800)))
(assert (zero? (time/parse "2006-01-02 15:04:05" "0001-01-01 00:00:00" "UTC")))
(assert (zero? (time/parse "2006-01-02 15:04:05" "0000-01-01 00:00:00" "UTC")))
(assert (zero? (time/parse "2006-01-02 15:04:05" "0000-00-00 00:00:00" "UTC")))
