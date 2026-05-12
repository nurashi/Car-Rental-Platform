"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { Car, BookOpen, User, LogOut, Menu, X } from "lucide-react";

export default function Navbar() {
  const router = useRouter();
  const pathname = usePathname();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    setIsLoggedIn(!!token);
  }, [pathname]);

  const handleLogout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setIsLoggedIn(false);
    router.push("/login");
  };

  const linkClass = (href: string) =>
    `flex items-center gap-1.5 text-sm font-medium transition-colors ${
      pathname === href
        ? "text-white"
        : "text-slate-300 hover:text-white"
    }`;

  return (
    <nav className="bg-slate-900 border-b border-slate-800 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <Link href="/" className="flex items-center gap-2 text-white font-bold text-xl">
            <div className="bg-primary-600 w-8 h-8 rounded-lg flex items-center justify-center">
              <Car size={18} />
            </div>
            RentCar
          </Link>

          <div className="hidden md:flex items-center gap-1">
            <Link href="/vehicles" className={linkClass("/vehicles") + " px-3 py-2 rounded-lg"}>
              <Car size={16} />
              Vehicles
            </Link>
            {isLoggedIn && (
              <>
                <Link href="/bookings" className={linkClass("/bookings") + " px-3 py-2 rounded-lg"}>
                  <BookOpen size={16} />
                  Bookings
                </Link>
                <Link href="/profile" className={linkClass("/profile") + " px-3 py-2 rounded-lg"}>
                  <User size={16} />
                  Profile
                </Link>
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-1.5 px-3 py-2 text-sm text-slate-300 hover:text-white transition-colors rounded-lg"
                >
                  <LogOut size={16} />
                  Logout
                </button>
              </>
            )}
            {!isLoggedIn && (
              <Link href="/login" className="btn btn-primary text-sm !px-4 !py-2 ml-2">
                Sign In
              </Link>
            )}
          </div>

          <button
            className="md:hidden text-white p-2"
            onClick={() => setMobileOpen(!mobileOpen)}
          >
            {mobileOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

        {mobileOpen && (
          <div className="md:hidden border-t border-slate-800 py-3 space-y-1 animate-slide-down">
            <Link href="/vehicles" className={linkClass("/vehicles") + " px-3 py-2.5 rounded-lg block"} onClick={() => setMobileOpen(false)}>
              <Car size={16} className="inline mr-2" />
              Vehicles
            </Link>
            {isLoggedIn && (
              <>
                <Link href="/bookings" className={linkClass("/bookings") + " px-3 py-2.5 rounded-lg block"} onClick={() => setMobileOpen(false)}>
                  <BookOpen size={16} className="inline mr-2" />
                  Bookings
                </Link>
                <Link href="/profile" className={linkClass("/profile") + " px-3 py-2.5 rounded-lg block"} onClick={() => setMobileOpen(false)}>
                  <User size={16} className="inline mr-2" />
                  Profile
                </Link>
                <button
                  onClick={() => { handleLogout(); setMobileOpen(false); }}
                  className="w-full text-left px-3 py-2.5 text-sm text-slate-300 hover:text-white transition-colors rounded-lg"
                >
                  <LogOut size={16} className="inline mr-2" />
                  Logout
                </button>
              </>
            )}
            {!isLoggedIn && (
              <Link href="/login" className="btn btn-primary block text-center mt-2" onClick={() => setMobileOpen(false)}>
                Sign In
              </Link>
            )}
          </div>
        )}
      </div>
    </nav>
  );
}
