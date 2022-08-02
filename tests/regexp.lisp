(let* 
     [re (regexp-compile "hello")
      loc (regexp-find-index re "012345hello!")]
  (assert (= (aget loc 0) 6))
  (assert (= (aget loc 1) 11))
  (assert (= "hello" (regexp-find re "ahellob")))
  (assert (regexp-match re "hello"))
  (assert (not (regexp-match re "hell"))))

(assert (str/contains?  (str (regexp-compile "hello")) "hello"))
(assert (= "(time/parse 1659438315220 'timestamp-ms)" (str (time/parse 1659438315220 'timestamp-ms))))
