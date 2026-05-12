"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import AuthGuard from "@/components/AuthGuard";
import { Calendar, DollarSign, Clock, XCircle, ChevronRight, MapPin } from "lucide-react";
import Link from "next/link";

interface Booking {
  id: string;
  vehicle_id: string;
  status: string;
  start_date: string;
  end_date: string;
  total_price: number;
  currency: string;
  payment_status: string;
}

export default function BookingsPage() {
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchBookings = async () => {
      const userStr = localStorage.getItem("user");
      if (!userStr) return;
      try {
        const user = JSON.parse(userStr);
        const userId = user.id || user.ID;
        const data = await api.bookings.list(userId);
        setBookings(data.bookings || []);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Failed to load bookings");
      } finally {
        setLoading(false);
      }
    };
    fetchBookings();
  }, []);

  const handleCancel = async (id: string) => {
    try {
      await api.bookings.cancel(id, "Cancelled by user");
      setBookings(bookings.filter((b) => b.id !== id));
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to cancel booking");
    }
  };

  const statusBadge = (status: string) => {
    const map: Record<string, string> = {
      active: "badge-success",
      confirmed: "badge-success",
      completed: "badge-info",
      pending: "badge-warning",
      cancelled: "badge-danger",
    };
    return `badge ${map[status] || "badge-neutral"}`;
  };

  const paymentBadge = (status: string) => {
    const map: Record<string, string> = {
      paid: "badge-success",
      unpaid: "badge-warning",
      refunded: "badge-info",
      partially_refunded: "badge-info",
    };
    return `badge text-[11px] ${map[status] || "badge-neutral"}`;
  };

  const formatDate = (d: string) => new Date(d).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  return (
    <AuthGuard>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12 animate-fade-in">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-10">
          <div>
            <h1 className="text-3xl font-bold text-slate-900">My Bookings</h1>
            <p className="text-slate-500 mt-1">{bookings.length} reservation{bookings.length !== 1 ? "s" : ""}</p>
          </div>
          <Link href="/vehicles" className="btn btn-primary">
            Book Another
            <ChevronRight size={16} />
          </Link>
        </div>

        {error && (
          <div className="mb-8 p-4 bg-red-50 border border-red-200 rounded-xl text-sm text-red-700 animate-slide-down">
            {error}
          </div>
        )}

        {loading ? (
          <div className="space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="card animate-pulse">
                <div className="flex justify-between">
                  <div className="space-y-3 flex-1">
                    <div className="h-5 bg-slate-200 rounded w-1/4" />
                    <div className="h-4 bg-slate-100 rounded w-1/3" />
                    <div className="h-4 bg-slate-100 rounded w-1/2" />
                  </div>
                  <div className="space-y-3">
                    <div className="h-6 bg-slate-200 rounded w-20" />
                    <div className="h-4 bg-slate-100 rounded w-16" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : bookings.length === 0 ? (
          <div className="text-center py-20">
            <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-slate-100 mb-6">
              <Calendar size={36} className="text-slate-400" />
            </div>
            <h3 className="text-lg font-semibold text-slate-700 mb-2">No bookings yet</h3>
            <p className="text-slate-500 mb-6">Start your journey by reserving a vehicle.</p>
            <Link href="/vehicles" className="btn btn-primary">
              Browse Vehicles
            </Link>
          </div>
        ) : (
          <div className="space-y-4">
            {bookings.map((b) => (
              <div key={b.id} className="card group hover:!shadow-md transition-all duration-300">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-2 flex-1">
                    <div className="flex items-center gap-2">
                      <span className={statusBadge(b.status)}>{b.status}</span>
                      <span className={paymentBadge(b.payment_status)}>{b.payment_status}</span>
                    </div>

                    <div className="flex items-center gap-2 text-sm text-slate-500">
                      <MapPin size={14} />
                      <span className="font-mono text-xs text-slate-400">#{b.id.slice(0, 8)}</span>
                    </div>

                    <div className="flex flex-wrap gap-x-6 gap-y-1">
                      <div className="flex items-center gap-1.5 text-sm text-slate-600">
                        <Calendar size={14} />
                        <span>{formatDate(b.start_date)} — {formatDate(b.end_date)}</span>
                      </div>
                      <div className="flex items-center gap-1.5 text-sm font-semibold text-slate-900">
                        <DollarSign size={14} />
                        <span>{b.total_price?.toFixed(2)} {b.currency}</span>
                      </div>
                    </div>

                    <div className="flex items-center gap-1.5 text-xs text-slate-400">
                      <Clock size={12} />
                      <span>{Math.max(1, Math.ceil((new Date(b.end_date).getTime() - new Date(b.start_date).getTime()) / 86400000))} day{Math.ceil((new Date(b.end_date).getTime() - new Date(b.start_date).getTime()) / 86400000) !== 1 ? "s" : ""}</span>
                    </div>
                  </div>

                  <div className="flex gap-2 sm:self-start">
                    {(b.status === "pending" || b.status === "confirmed") && (
                      <button
                        className="btn btn-danger !px-4 !py-2 !text-xs"
                        onClick={() => handleCancel(b.id)}
                      >
                        <XCircle size={14} />
                        Cancel
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </AuthGuard>
  );
}
