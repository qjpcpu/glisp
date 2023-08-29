(assert (= 1 (hget (let [args "a"] (filter #(= args (car %1))  {"a" 1 "b" 2})) "a")))
