import { TableProps as ElTableProps } from 'element-plus'
export interface TableColumn {
  field: string
  label?: string
  type?: string
  /**
   * Whether to hide
   */
  hidden?: boolean
  children?: TableColumn[]
  slots?: {
    default?: (...args: any[]) => JSX.Element | JSX.Element[] | null
    header?: (...args: any[]) => JSX.Element | null
  }
  index?: number | ((index: number) => number)
  columnKey?: string
  width?: string | number
  minWidth?: string | number
  fixed?: boolean | 'left' | 'right'
  renderHeader?: (...args: any[]) => JSX.Element | null
  // sortable?: boolean
  sortMethod?: (...args: any[]) => number
  sortBy?: string | string[] | ((...args: any[]) => string | string[])
  sortOrders?: (string | null)[]
  resizable?: boolean
  formatter?: (...args: any[]) => any
  showOverflowTooltip?: boolean
  align?: 'left' | 'center' | 'right'
  headerAlign?: 'left' | 'center' | 'right'
  className?: string
  labelClassName?: string
  selectable?: (...args: any[]) => boolean
  reserveSelection?: boolean
  filters?: Array<{ text: string; value: string }>
  filterPlacement?: string
  filterMultiple?: boolean
  filterMethod?: (...args: any[]) => boolean
  filteredValue?: string[]
  [key: string]: any
}

export interface TableSlotDefault {
  row: Recordable
  column: TableColumn
  $index: number
  [key: string]: any
}

export interface Pagination {
  small?: boolean
  background?: boolean
  pageSize?: number
  defaultPageSize?: number
  total?: number
  pageCount?: number
  pagerCount?: number
  currentPage?: number
  defaultCurrentPage?: number
  layout?: string
  pageSizes?: number[]
  popperClass?: string
  prevText?: string
  nextText?: string
  disabled?: boolean
  hideOnSinglePage?: boolean
}

export interface TableSetProps {
  field: string
  path: string
  value: any
}

export interface TableProps extends Omit<Partial<ElTableProps<any[]>>, 'data'> {
  pageSize?: number
  currentPage?: number
  showAction?: boolean
  // Whether to hide all overflow, priority is lower than showOverflowTooltip in schema,
  showOverflowTooltip?: boolean
  // Table header
  columns?: TableColumn[]
  // Whether to show pagination
  pagination?: Pagination | undefined
  // Only valid for columns with type=selection, Boolean type; when true, previously selected data is kept after data updates (row-key must be specified)
  reserveSelection?: boolean
  // Loading state
  loading?: boolean
  // Whether to overlay the index
  reserveIndex?: boolean
  // Alignment
  align?: 'left' | 'center' | 'right'
  // Header alignment
  headerAlign?: 'left' | 'center' | 'right'
  preview?: string[]
  sortable?: boolean
  data?: Recordable
}
