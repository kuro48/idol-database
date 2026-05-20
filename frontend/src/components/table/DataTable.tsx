import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from '@tanstack/react-table'
import { Skeleton } from '../ui/Skeleton'
import { EmptyState } from '../ui/EmptyState'
import styles from './data-table.module.css'

const SKELETON_ROW_COUNT = 8

interface DataTableProps<TData> {
  columns: ColumnDef<TData, unknown>[]
  data: TData[]
  isLoading?: boolean
  emptyMessage?: string
}

export function DataTable<TData>({
  columns,
  data,
  isLoading = false,
  emptyMessage = 'No results found.',
}: DataTableProps<TData>) {
  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <div className={styles.wrapper}>
      <table className={styles.table}>
        <thead className={styles.thead}>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th key={header.id} className={styles.th}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext(),
                      )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody className={styles.tbody}>
          {isLoading
            ? Array.from({ length: SKELETON_ROW_COUNT }).map((_, i) => (
                <tr key={i} className={styles.tr}>
                  {columns.map((_, j) => (
                    <td key={j} className={styles.td}>
                      <Skeleton height="1rem" />
                    </td>
                  ))}
                </tr>
              ))
            : table.getRowModel().rows.length === 0
              ? (
                <tr>
                  <td colSpan={columns.length}>
                    <EmptyState title={emptyMessage} />
                  </td>
                </tr>
              )
              : table.getRowModel().rows.map((row) => (
                  <tr key={row.id} className={styles.tr}>
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className={styles.td}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </td>
                    ))}
                  </tr>
                ))}
        </tbody>
      </table>
    </div>
  )
}
