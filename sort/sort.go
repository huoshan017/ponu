package sort

type Elem interface {
	Less(e Elem) bool
	Equal(e Elem) bool
}

func less(e1, e2 Elem) bool {
	return e1.Less(e2)
}

func greater(e1, e2 Elem) bool {
	return e2.Less(e1)
}

func less_equal(e1, e2 Elem) bool {
	return e1.Less(e2) || e1.Equal(e2)
}

func greater_equal(e1, e2 Elem) bool {
	return !e1.Less(e2)
}

// ======================== quick sort =========================
func qsort_partition(arr []Elem, low, high int, desc bool) int {
	pivot := arr[low]
	for low < high {
		for low < high && ((!desc && less_equal(pivot, arr[high])) || (desc && less_equal(arr[high], pivot))) {
			high -= 1
		}
		if low < high {
			arr[low] = arr[high]
		}
		for low < high && ((!desc && less_equal(arr[low], pivot)) || (desc && less_equal(pivot, arr[low]))) {
			low += 1
		}
		if low < high {
			arr[high] = arr[low]
		}
	}
	arr[low] = pivot
	return low
}

func qsort(arr []Elem, low, high int, desc bool) {
	if low < high {
		pivot := qsort_partition(arr, low, high, desc)
		qsort(arr, low, pivot-1, desc)
		qsort(arr, pivot+1, high, desc)
	}
}

func QSort(arr []Elem, desc bool) {
	qsort(arr, 0, len(arr)-1, desc)
}

// ======================= heap sort ========================
func heap_adjust_max(arr []Elem, s, m int) {
	if m > len(arr)-1 {
		m = len(arr) - 1
	}

	var max int
	for {
		left := 2*s + 1
		if left > m {
			break
		}
		right := 2*s + 2
		max = left
		if right <= m {
			if arr[left].Less(arr[right]) {
				max = right
			}
		}
		if !arr[s].Less(arr[max]) {
			break
		}
		arr[s], arr[max] = arr[max], arr[s]
		s = max
	}
}

func heap_adjust_min(arr []Elem, s, m int) {
	if m > len(arr)-1 {
		m = len(arr) - 1
	}

	var min int
	for {
		left := 2*s + 1
		if left > m {
			break
		}
		right := 2*s + 2
		min = left
		if right <= m {
			if arr[right].Less(arr[left]) {
				min = right
			}
		}
		if !arr[min].Less(arr[s]) {
			break
		}
		arr[s], arr[min] = arr[min], arr[s]
		s = min
	}
}

func HeapSort(arr []Elem, desc bool) {
	if !desc {
		for i := len(arr)/2 - 1; i >= 0; i-- {
			heap_adjust_max(arr, i, len(arr)-1)
		}
		for i := len(arr) - 1; i > 0; i-- {
			arr[0], arr[i] = arr[i], arr[0]
			heap_adjust_max(arr, 0, i-1)
		}
	} else {
		for i := len(arr)/2 - 1; i >= 0; i-- {
			heap_adjust_min(arr, i, len(arr)-1)
		}
		for i := len(arr) - 1; i > 0; i-- {
			arr[0], arr[i] = arr[i], arr[0]
			heap_adjust_min(arr, 0, i-1)
		}
	}
}

// ============================== merge sort =============================
func MergeSort(arr []Elem, desc bool) (out_arr []Elem) {
	var arr2 []Elem
	var l = len(arr)
	if l > 2 {
		arr2 = make([]Elem, l)
	}

	var b bool
	for s := 1; s <= l; s = s * 2 {
		if s == 1 {
			merge_first(arr, desc)
			out_arr = arr
		} else {
			if !b {
				merge(arr, s, arr2, desc)
			} else {
				merge(arr2, s, arr, desc)
			}
			b = !b
		}
	}
	if !b {
		out_arr = arr
	} else {
		out_arr = arr2
	}
	return out_arr
}

// when step is 2
func merge_first(arr []Elem, desc bool) {
	l := len(arr)
	for s := 0; s < l; s = s + 2 {
		if s+1 >= l {
			break
		}
		if (!desc && arr[s+1].Less(arr[s])) || (desc && arr[s].Less(arr[s+1])) {
			arr[s], arr[s+1] = arr[s+1], arr[s]
		}
	}
}

func merge(arr []Elem, step int, out_arr []Elem, desc bool) {
	var (
		s, i, j, n, imax, jmax int
	)
	l := len(arr)
	for s = 0; s < l; s = s + 2*step {
		if l <= s+step {
			break
		}

		i = s
		j = s + step
		n = s
		imax = s + step
		if s+step+step > l {
			jmax = l
		} else {
			jmax = s + step + step
		}
		for i < imax && j < jmax {
			if (!desc && less_equal(arr[i], arr[j])) || (desc && less_equal(arr[j], arr[i])) {
				out_arr[n] = arr[i]
				i += 1
			} else {
				out_arr[n] = arr[j]
				j += 1
			}
			n += 1
		}
		if i < imax {
			for ; i < imax; i++ {
				out_arr[n] = arr[i]
				n += 1
			}
		} else if j < jmax {
			for ; j < jmax; j++ {
				out_arr[n] = arr[j]
				n += 1
			}
		}
	}
}
