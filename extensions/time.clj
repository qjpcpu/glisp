(defn time/format-beijing [tm]
  "Usage: (time/format-beijing)

Return time string in beijing."
  (time/format tm "2006-01-02 15:04:05" "Asia/Shanghai"))

(defn time/format-utc [tm]
  "Usage: (time/format-utc)

Return time string in utc."
  (time/format tm "2006-01-02 15:04:05" "UTC"))

(defn time/parse-beijing [tm]
  "Usage: (time/parse-beijing)

Parse time string of format 2006-01-02 15:04:05 in beijing."
  (time/parse tm "2006-01-02 15:04:05" "Asia/Shanghai"))

(defn time/parse-utc [tm]
  "Usage: (time/parse-utc)

Parse time string of format 2006-01-02 15:04:05 in utc."
  (time/parse tm "2006-01-02 15:04:05" "UTC"))
