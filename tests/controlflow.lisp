(assert (not (and false false)))
(assert (not (and true false)))
(assert (not (and false true)))
(assert (and true true))
(assert (= 0 (and 1 0)))
(assert (= 1 (and 0 1)))
(assert (= 2 (and 1 2)))

(assert (not (or false false)))
(assert (or false true))
(assert (or true false))
(assert (or true true))
(assert (= 0 (or 0 1)))
(assert (= 1 (or 1 0)))
(assert (= 1 (or 1 2)))
(assert (= "" (string (or (bytes "") (bytes "a")))))
(assert (= "a" (string(and (bytes "") (bytes "a")))))
(assert (= "" (or "" "a")))
(assert (= "a" (and "" "a")))

(assert (= 'a
  (cond true 'a 'b)))

(assert (= 'b
  (cond false 'a 'b)))

(assert (= 'c
  (cond
    false 'a
    false 'b
          'c)))
