(defn list/complement [a b]
  "Usage: (list/complement a b)

Return coll which elements belongs to a but not belongs to b.
Result coll = a \\ b."
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
       (filter (fn [e] (not (exist? h e))) a)))

(defn list/intersect [a b]
  "Usage: (list/intersect a b)
Return coll which elements belongs to a and b."
  (let [h (foldl (fn [e acc] (hset! acc e '()) acc) {} b)]
       (filter (fn [e] (exist? h e)) a)))

(defn list/uniq [a]
  "Usage: (list/uniq a)
Drop duplicate elements of coll a."
  (let [h {}]
       (filter (fn [e]
                   (cond (exist? h e) false
                         (begin (hset! h e '()) true))) a)))

(defn list/union [a b]
  "Usage: (list/union a b)
Returns coll contains all elements of a and b."
  (concat a b))
