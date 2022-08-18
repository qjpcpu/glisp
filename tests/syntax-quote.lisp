(def a 7)
(def x `[{'g ({'b [~a]})}])
(assert (= (str x) "[(hash (quote g) ((hash (quote b) [7])))]"))

(assert (= '(hash "a" 1) (syntax-quote {"a" 1})))
(assert (= "hash" (typestr (eval (syntax-quote {"a" 1})))))

