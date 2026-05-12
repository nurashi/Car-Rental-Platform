"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api";
import AuthGuard from "@/components/AuthGuard";
import { Car, Calendar, DollarSign, ArrowLeft, CheckCircle, Clock, MapPin, Gauge } from "lucide-react";

interface Vehicle {
  id: string;
  make: string;
  model: string;
  year: number;
  license_plate: string;
  status: string;
  mileage: number;
  current_location_id?: string;
}

export default function VehicleDetailPage() {
  const params = useParams();
  const router = useRouter();
  const [vehicle, setVehicle] = useState<Vehicle | null>(null);
  const [loading, setLoading] = useState(true);
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [price, setPrice] = useState<Record<string, unknown> | null>(null);
  const [priceLoading, setPriceLoading] = useState(false);
  const [error, setError] = useState("");
  const [bookingSuccess, setBookingSuccess] = useState(false);
  const [bookingLoading, setBookingLoading] = useState(false);

  useEffect(() => {
    const fetchVehicle = async () => {
      try {
        const data = await api.vehicles.get(params.id as string);
        setVehicle(data.vehicle);
      } catch {
        setError("Vehicle not found");
      } finally {
        setLoading(false);
      }
    };
    fetchVehicle();
  }, [params.id]);

  const handleCalculatePrice = async () => {
    if (!startDate || !endDate) return;
    setPriceLoading(true);
    setError("");
    try {
      const data = await api.bookings.calculatePrice(params.id as string, startDate + "T00:00:00Z", endDate + "T00:00:00Z");
      setPrice(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to calculate price");
    } finally {
      setPriceLoading(false);
    }
  };

  const handleBooking = async () => {
    setBookingLoading(true);
    setError("");
    try {
      await api.bookings.create({
        vehicle_id: params.id as string,
        start_date: startDate + "T00:00:00Z",
        end_date: endDate + "T00:00:00Z",
      });
      setBookingSuccess(true);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to create booking");
    } finally {
      setBookingLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-12 animate-pulse">
        <div className="h-8 bg-slate-200 rounded w-1/3 mb-4" />
        <div className="h-4 bg-slate-100 rounded w-1/2 mb-8" />
        <div className="card h-64" />
      </div>
    );
  }

  if (bookingSuccess) {
    return (
      <div className="max-w-lg mx-auto px-4 py-20 text-center animate-scale-in">
        <div className="card !p-10">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-emerald-100 mb-6">
            <CheckCircle size={40} className="text-emerald-600" />
          </div>
          <h2 className="text-2xl font-bold text-slate-900 mb-2">Booking Confirmed!</h2>
          <p className="text-slate-500 mb-8">Your reservation has been placed successfully.</p>
          <button className="btn btn-primary !px-8 !py-3" onClick={() => router.push("/bookings")}>
            View My Bookings
          </button>
        </div>
      </div>
    );
  }

  if (!vehicle) {
    return (
      <div className="max-w-lg mx-auto px-4 py-20 text-center">
        <div className="card !p-10">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-red-100 mb-6">
            <Car size={36} className="text-red-500" />
          </div>
          <h2 className="text-xl font-bold text-slate-900 mb-2">Vehicle Not Found</h2>
          <p className="text-slate-500 mb-6">{error || "This vehicle may have been removed."}</p>
          <button className="btn btn-primary" onClick={() => router.push("/vehicles")}>
            <ArrowLeft size={16} />
            Back to Vehicles
          </button>
        </div>
      </div>
    );
  }

  const isAvailable = vehicle.status === "available";
  const today = new Date().toISOString().split("T")[0];

  return (
    <AuthGuard>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12 animate-fade-in">
        <button
          onClick={() => router.push("/vehicles")}
          className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-700 mb-8 transition-colors"
        >
          <ArrowLeft size={16} />
          Back to vehicles
        </button>

        <div className="grid lg:grid-cols-5 gap-8">
          <div className="lg:col-span-3">
            <div className="card !p-8">
              <div className="flex items-start justify-between mb-6">
                <div>
                  <h1 className="text-2xl sm:text-3xl font-bold text-slate-900">
                    {vehicle.make} {vehicle.model}
                  </h1>
                  <p className="text-slate-500 mt-1">{vehicle.year}</p>
                </div>
                <span className={`badge text-sm !px-4 !py-1.5 ${isAvailable ? "badge-success" : vehicle.status === "maintenance" ? "badge-warning" : "badge-danger"}`}>
                  {vehicle.status}
                </span>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
                <div className="flex items-center gap-2 p-3 bg-slate-50 rounded-xl">
                  <MapPin size={18} className="text-slate-400" />
                  <div>
                    <div className="text-xs text-slate-500">Plate</div>
                    <div className="text-sm font-semibold text-slate-900">{vehicle.license_plate}</div>
                  </div>
                </div>
                <div className="flex items-center gap-2 p-3 bg-slate-50 rounded-xl">
                  <Gauge size={18} className="text-slate-400" />
                  <div>
                    <div className="text-xs text-slate-500">Mileage</div>
                    <div className="text-sm font-semibold text-slate-900">{vehicle.mileage > 0 ? vehicle.mileage.toLocaleString() : "0"} km</div>
                  </div>
                </div>
                <div className="flex items-center gap-2 p-3 bg-slate-50 rounded-xl">
                  <Clock size={18} className="text-slate-400" />
                  <div>
                    <div className="text-xs text-slate-500">Year</div>
                    <div className="text-sm font-semibold text-slate-900">{vehicle.year}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="lg:col-span-2">
            <div className="card !p-8 sticky top-24">
              <h3 className="text-lg font-semibold text-slate-900 mb-6 flex items-center gap-2">
                <Calendar size={20} className="text-primary-600" />
                Book This Vehicle
              </h3>

              {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-xl text-sm text-red-700 animate-slide-down">
                  {error}
                </div>
              )}

              {!isAvailable && (
                <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-xl text-sm text-amber-700">
                  This vehicle is currently {vehicle.status} and cannot be booked.
                </div>
              )}

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1.5">Start Date</label>
                  <input
                    type="date"
                    className="input"
                    min={today}
                    value={startDate}
                    onChange={(e) => { setStartDate(e.target.value); setPrice(null); }}
                    disabled={!isAvailable}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-1.5">End Date</label>
                  <input
                    type="date"
                    className="input"
                    min={startDate || today}
                    value={endDate}
                    onChange={(e) => { setEndDate(e.target.value); setPrice(null); }}
                    disabled={!isAvailable || !startDate}
                  />
                </div>

                <button
                  className="btn btn-secondary !w-full"
                  onClick={handleCalculatePrice}
                  disabled={!startDate || !endDate || !isAvailable || priceLoading}
                >
                  {priceLoading ? (
                    <div className="w-4 h-4 border-2 border-slate-400 border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <DollarSign size={18} />
                  )}
                  Calculate Price
                </button>

                {price && (
                  <div className="p-4 bg-gradient-to-br from-primary-50 to-blue-50 rounded-xl space-y-2 animate-slide-up">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-slate-600">Total Price</span>
                      <span className="text-2xl font-bold text-primary-700">
                        ${(price.total_price as number)?.toFixed(2)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm text-slate-500">
                      <span>Duration</span>
                      <span>{(price.rental_days as number) || 0} days</span>
                    </div>
                    <div className="flex items-center justify-between text-sm text-slate-500">
                      <span>Currency</span>
                      <span>{price.currency as string || "USD"}</span>
                    </div>
                  </div>
                )}

                <button
                  className="btn btn-primary !w-full !py-3"
                  onClick={handleBooking}
                  disabled={!startDate || !endDate || !price || !isAvailable || bookingLoading}
                >
                  {bookingLoading ? (
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  ) : (
                    <>
                      <CheckCircle size={18} />
                      Book Now
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AuthGuard>
  );
}
