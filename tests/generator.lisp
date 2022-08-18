(defmac test-gh [h] `{"a" ~h})
(def h (test-gh {"b" 1}))
(assert (= 1 (json/query h "a.b")))
