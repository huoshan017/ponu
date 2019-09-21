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
func qsort_partition(arr []Elem, low, high int) int {
	pivot := arr[low]
	for low < high {
		for low < high && less_equal(pivot, arr[high]) {
			high -= 1
		}
		if low < high {
			arr[low] = arr[high]
		}
		for low < high && less_equal(arr[low], pivot) {
			low += 1
		}
		if low < high {
			arr[high] = arr[low]
		}
	}
	arr[low] = pivot
	return low
}

func qsort(arr []Elem, low, high int) {
	if low < high {
		pivot := qsort_partition(arr, low, high)
		qsort(arr, low, pivot-1)
		qsort(arr, pivot+1, high)
	}
}

func QSort(arr []Elem) {
	qsort(arr, 0, len(arr)-1)
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

	var max int
	for {
		left := 2*s + 1
		if left > m {
			break
		}
		right := 2*s + 2
		max = left
		if right <= m {
			if arr[right].Less(arr[left]) {
				max = right
			}
		}
		if !arr[max].Less(arr[s]) {
			break
		}
		arr[s], arr[max] = arr[max], arr[s]
		s = max
	}
}

func HeapSortMax(arr []Elem) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		heap_adjust_max(arr, i, len(arr)-1)
	}
	for i := len(arr) - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heap_adjust_max(arr, 0, i-1)
	}
}

func HeapSortMin(arr []Elem) {
	for i := len(arr)/2 - 1; i >= 0; i-- {
		heap_adjust_min(arr, i, len(arr)-1)
	}
	for i := len(arr) - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heap_adjust_min(arr, 0, i-1)
	}
}
