(defn json/__q_hash [js indent]
  (str/join (concat [(concat indent "{")]
                 (foldl (fn [k v acc]
                                   (cond (array? v) (append acc (concat indent "    " (json/stringify k) ": " (sprintf "[<len=%v>]" (len v))))
                                         (hash? v) (append acc (concat indent "    " (json/stringify k) ": {" (str/join  (foldl (fn [k1 v1 acc1] (append acc1 (json/stringify k1))) [] v) ",") "}"))
                                         (string? v) (append acc (concat indent "    " (json/stringify k) ": " (json/__q_str v)))
                                         (append acc (concat indent "    " (json/stringify k) ": " (json/stringify v))))) [] js)
                 [(concat indent "}")]) "\n"))

(defn json/__q_str [js]
  (cond (> (len js) 64) (json/stringify (sprintf "%s...<len=%v>" (slice js 0 64) (len js))) (json/stringify js)))

(defn json/q [js & args]
  "Usage: (json/q hash & args)
(json/q hash path) ; show data summary by path
(json/q hash true) ; show full json data
(json/q hash path true) ; show full json data by path

Query json object by path."
  (cond (nil? args)
          (cond (nil? js) (println (json/stringify js))
                (array? js) (println (str/join (concat ["["]
                                           (append (map (fn [e]
                                              (cond (array? e) (sprintf "    [<len=%v>]" (len e))
                                                    (hash? e) (json/__q_hash e "    ")
                                                    (string? e) (concat "    " (json/__q_str e))
                                                    (concat "    " (json/stringify e)))
                                              ) (slice js 0 3))
                        "    ......"
                        (sprintf "    <len=%v>" (len js))
                        "]"))
                          "\n"))
                (hash? js)  (println (json/__q_hash js ""))
                (string? js) (println (json/__q_str js))
                (println (json/stringify js)))
        (bool? (car args)) (println (json/stringify js (car args)))
        (nil? (cdr args)) (json/q (json/query js (car args)))
        (json/q (json/query js (car args)) (car (cdr args)))))
