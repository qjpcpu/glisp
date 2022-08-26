(defn json/__q_hash [js indent]
  (str/join (concat [(concat indent "{")]
                 (foldl (fn [k v acc]
                                   (cond (array? v) (append acc (concat indent "    " (json/stringify k) ": " (sprintf "[<len=%v>]" (len v))))
                                         (hash? v) (append acc (concat indent "    " (json/stringify k) ": {" (str/join  (foldl (fn [k1 v1 acc1] (append acc1 (json/stringify k1))) [] v) ",") "}"))
                                         (append acc (concat indent "    " (json/stringify k) ": " (json/stringify v))))) [] js)
                 [(concat indent "}")]) "\n"))

(defn json/q [js & args]
  (cond (null? args)
          (cond (null? js) (println (json/stringify js))
                (array? js) (println (str/join (concat ["["]
                                           (append (map (fn [e]
                                                            (cond (array? e) (sprintf "    [<len=%v>]" (len e))
                                                                  (hash? e) (json/__q_hash e "    ")
                                                                  (concat "    " (json/stringify e)))
                                                            ) (slice js 0 3))
                        "    ......"
                        (sprintf "    <total=%v>" (len js))
                        "]"))
                          "\n"))
                (hash? js)  (println (json/__q_hash js ""))
                (println (json/stringify js)))
        (bool? (car args)) (println (json/stringify js (car args)))
        (null? (cdr args)) (json/q (json/query js (car args)))
        (json/q (json/query js (car args)) (car (cdr args)))))
