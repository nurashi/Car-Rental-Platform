"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { Car, MapPin, Calendar, ChevronLeft, ChevronRight, Search } from "lucide-react";
import Link from "next/link";

interface Vehicle {
  id: string;
  make: string;
  model: string;
  year: number;
  license_plate: string;
  status: string;
  mileage: number;
}

export default function VehiclesPage() {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);

  const fetchVehicles = async () => {
    setLoading(true);
    try {
      const data = await api.vehicles.list(page, 10);
      setVehicles(data.vehicles || []);
      setTotal(data.total || 0);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to load vehicles");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchVehicles();
  }, [page]);

  const totalPages = Math.ceil(total / 10);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 animate-fade-in">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-10">
        <div>
          <h1 className="text-3xl font-bold text-slate-900">Available Vehicles</h1>
          <p className="text-slate-500 mt-1">{total} vehicles in our fleet</p>
        </div>
        <div className="relative">
          <Search size={18} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
          <input className="input !pl-11 !w-64" placeholder="Search vehicles..." disabled />
        </div>
      </div>

      {error && (
        <div className="mb-8 p-4 bg-red-50 border border-red-200 rounded-xl text-sm text-red-700 animate-slide-down">
          {error}
        </div>
      )}

      {loading ? (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="card animate-pulse">
              <div className="h-5 bg-slate-200 rounded w-2/3 mb-3" />
              <div className="h-4 bg-slate-100 rounded w-1/2 mb-4" />
              <div className="flex justify-between">
                <div className="h-6 bg-slate-100 rounded w-20" />
                <div className="h-9 bg-slate-200 rounded w-28" />
              </div>
            </div>
          ))}
        </div>
      ) : vehicles.length === 0 ? (
        <div className="text-center py-20">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-slate-100 mb-6">
            <Car size={36} className="text-slate-400" />
          </div>
          <h3 className="text-lg font-semibold text-slate-700 mb-2">No vehicles found</h3>
          <p className="text-slate-500">Check back later for new additions to our fleet.</p>
        </div>
      ) : (
        <>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {vehicles.map((v, i) => (
              <div
                key={v.id}
                className="card group hover:!shadow-lg hover:-translate-y-1 transition-all duration-300"
                style={{ animationDelay: `${i * 50}ms` }}
              >
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="text-lg font-semibold text-slate-900 group-hover:text-primary-600 transition-colors">
                      {v.make} {v.model}
                    </h3>
                    <p className="text-sm text-slate-500">{v.year}</p>
                  </div>
                  <span className={`badge ${v.status === "available" ? "badge-success" : v.status === "maintenance" ? "badge-warning" : "badge-danger"}`}>
                    {v.status}
                  </span>
                </div>

                <div className="space-y-2 mb-4">
                  <div className="flex items-center gap-2 text-sm text-slate-500">
                    <MapPin size={14} />
                    <span>Plate: {v.license_plate}</span>
                  </div>
                  <div className="flex items-center gap-2 text-sm text-slate-500">
                    <Calendar size={14} />
                    <span>{v.mileage > 0 ? v.mileage.toLocaleString() : "0"} km</span>
                  </div>
                </div>

                <Link
                  href={`/vehicles/${v.id}`}
                  className="btn btn-primary !w-full group/btn"
                >
                  View & Book
                  <ChevronRight size={16} className="group-hover/btn:translate-x-0.5 transition-transform" />
                </Link>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-3 mt-10">
              <button
                className="btn btn-secondary !px-4"
                disabled={page <= 1}
                onClick={() => setPage(page - 1)}
              >
                <ChevronLeft size={16} />
                Previous
              </button>
              <div className="flex items-center gap-1">
                {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={`w-9 h-9 rounded-lg text-sm font-medium transition-colors ${
                      p === page
                        ? "bg-primary-600 text-white shadow-sm"
                        : "text-slate-600 hover:bg-slate-100"
                    }`}
                  >
                    {p}
                  </button>
                ))}
              </div>
              <button
                className="btn btn-secondary !px-4"
                disabled={page >= totalPages}
                onClick={() => setPage(page + 1)}
              >
                Next
                <ChevronRight size={16} />
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
