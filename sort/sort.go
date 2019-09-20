package sort

type Elem interface {
	Less(e Elem) bool
	Equal(e Elem) bool
}

func qsort_partition(arr []Elem, low, high int) int {
	pivot := arr[low]
	for low < high {
		for low < high && (pivot.Less(arr[high]) || pivot.Equal(arr[high])) {
			high -= 1
		}
		if low < high {
			arr[low] = arr[high]
		}
		for low < high && (arr[low].Less(pivot) || arr[low].Equal(pivot)) {
			low += 1
		}
		if low < high {
			arr[high] = arr[low]
		}
	}
	arr[low] = pivot
	return low
}

func QSort(arr []Elem, low, high int) {
	if low < high {
		pivot := qsort_partition(arr, low, high)
		QSort(arr, low, pivot-1)
		QSort(arr, pivot+1, high)
	}
}
