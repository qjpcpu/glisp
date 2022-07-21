(assert (= '(1 2 3) (cons 1 '(2 3))))

(assert (= 1 (car '(1 2 3))))

(assert (= '(2 3) (cdr '(1 2 3))))

(assert (= 2 (car (cdr '(1 2 3)))))

(let [a 3]
  (assert (= '(0 3) (list 0 a))))

(assert (= '(1 2 4 5) (concat '(1 2) '(4 5))))
(assert (= '(1 2 4 5 6) (concat '(1 2) '() '(4 5) '(6))))
(assert (= '(1 2 3 4) (concat '() '(1 2) '(3 4))))
(assert (= '(1 2) (concat '() '(1 2))))

; test not-list pairs
(assert (= '(1 . 2) (cons 1 2)))
(assert (= 2 (cdr '(1 . 2))))
(assert (not (list? '(1 . 2))))
(assert (list? '()))
(assert (list? '(1 2 3)))
(assert (empty? '()))
(assert (not (empty? '(1 2))))
