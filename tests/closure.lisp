;; (defn foldl [lst fun acc]
;;     (cond
;;         (empty? lst) acc
;;         (foldl (cdr lst) fun (fun (car lst) acc))
;; 		))
        
;; (defn filter [lst fun]
;;     (foldl (fn [x l]
;;                 (cond
;;                     (fun x) (append l x)
;;                     l))
;;             [] lst)
;; )

(defn g-map [fun lst]
	(foldl (fn [x l] (fun x)) [] lst))

(defn even? [x]
    (cond
        (number? x)
            (cond
                (float? x) false
                (= (bit-and x 1) 0))
            false
            ))

(defn newStore [items]
	(let* [
		idx [-1]
		fun (fn []
				(cond
					(< (+ (aget idx 0) 1) (len items))
						(begin
							(aset! idx 0 (+ (aget idx 0) 1))
							(aget items (aget idx 0))
						)
					()))
			] fun))


(def evens (newStore [2 4 6 8 10]))

(map (fn [x]
		(assert (= x (evens))))
    (filter even? [1 2 3 4 5 6 7 8 9 10])
)

(def s (newStore [10 9 8 7 6 5 4 3 2 1]))
(g-map (fn [x] (assert (= x (s)))) [10 9 8 7 6 5 4 3 2 1])
(def s (newStore [10 9 8 7 6 5 4 3 2 1]))
(map (fn [x] (assert (= x (s)))) [10 9 8 7 6 5 4 3 2 1])

(def decending (newStore [10 9 8 7 6 5 4 3 2 1]))


(def s (newStore [10 9 8 7 6 5 4 3 2 1]))

(defn drainStore []
	(let [ v (s) ]
		(cond
			(empty? v) (assert (= (decending) ()))
			((begin
				(assert (= (decending) v))
				(drainStore)))))
	)
		

(drainStore)

(def a 1)
(defn enclosure []
  (let [func1 (fn [thunk] (let [a 2] (thunk)))
         func2 (fn [] (assert (= a 1)))]
    (func1 func2)))

(enclosure)
