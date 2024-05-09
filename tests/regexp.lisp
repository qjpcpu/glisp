(let* 
     [re (regexp/compile "hello")
      loc (regexp/find-index re "012345hello!")]
  (assert (= (aget loc 0) 6))
  (assert (= (aget loc 1) 11))
  (assert (= "hello" (regexp/find re "ahellob")))
  (assert (regexp/match re "hello"))
  (assert (not (regexp/match re "hell"))))

(assert (str/contains?  (string (regexp/compile "hello")) "hello"))
(assert (= "(time/parse 1659438315220 'timestamp-ms)" (sexp-str (time/parse 1659438315220 'timestamp-ms))))


(let*
     [re  "hello"
      loc (regexp/find-index re "012345hello!")]
  (assert (= (aget loc 0) 6))
  (assert (= (aget loc 1) 11))
  (assert (= "hello" (regexp/find re "ahellob")))
  (assert (regexp/match re "hello"))
  (assert (not (regexp/match re "hell"))))


(assert (= (regexp/replace (regexp/compile "a(x*)b") "-ab-axxb-" "T") "-T-T-"))
(assert (= (regexp/replace "a(x*)b" "-ab-axxb-" "T") "-T-T-"))

(assert (= (regexp/replace (regexp/compile "a(x*)b") "-ab-axxb-" "$1W") "---"))
(assert (= (regexp/replace "a(x*)b" "-ab-axxb-" "$1W") "---"))

(assert (= "regexp" (type (regexp/compile "[0-9]"))))

;; test cache
(->> (range 200) (map #(regexp/compile (string %))) (realize))
