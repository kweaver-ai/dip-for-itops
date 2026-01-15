import { useEffect, useState } from "react";

export const useTableFilter = (opts: {
  level: any[];
  status: any[];
}) => {
  const { level, status } = opts;
  const [filterLevel, setFilterLevel] = useState<any[]>([]);
  const [filterStatus, setFilterStatus] = useState<any[]>([]);

  useEffect(() => {
    setFilterLevel(level);
  }, [level]);

  useEffect(() => {
    setFilterStatus(status);
  }, [status]);

  return {
    filterLevel,
    filterStatus,
    setFilterLevel,
    setFilterStatus,
  }
}
