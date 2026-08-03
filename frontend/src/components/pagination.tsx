import { Pagination as AntPagination } from "antd";

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (size: number) => void;
}

export function Pagination({
  currentPage,
  total,
  pageSize,
  onPageChange,
  onPageSizeChange,
}: PaginationProps) {
  return (
    <div className="flex justify-end py-4">
      <AntPagination
        current={currentPage}
        total={total}
        pageSize={pageSize}
        showSizeChanger={Boolean(onPageSizeChange)}
        pageSizeOptions={[10, 20, 50, 100]}
        showTotal={(count, range) => `显示 ${range[0]} - ${range[1]} 条，共 ${count} 条`}
        onChange={(page, size) => {
          if (size !== pageSize) onPageSizeChange?.(size);
          else onPageChange(page);
        }}
      />
    </div>
  );
}
