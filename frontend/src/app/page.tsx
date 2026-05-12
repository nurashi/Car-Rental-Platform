"use client";

import Link from "next/link";
import { Car, Calendar, Search, Shield, Zap, ArrowRight } from "lucide-react";

export default function Home() {
  return (
    <div className="animate-fade-in">
      <section className="relative overflow-hidden bg-gradient-to-br from-slate-900 via-slate-800 to-primary-950 text-white">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-primary-900/20 via-transparent to-transparent" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32 lg:py-40">
          <div className="max-w-3xl">
            <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm rounded-full px-4 py-1.5 text-sm mb-8">
              <Zap size={14} className="text-primary-400" />
              Premium Fleet Available Now
            </div>
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight mb-6">
              Find Your{" "}
              <span className="bg-gradient-to-r from-primary-400 to-blue-400 bg-clip-text text-transparent">
                Perfect Ride
              </span>
            </h1>
            <p className="text-lg sm:text-xl text-slate-300 max-w-2xl mb-10 leading-relaxed">
              Professional car rental with a modern fleet. Transparent pricing, instant booking, and vehicles you can trust.
            </p>
            <div className="flex flex-wrap gap-4">
              <Link href="/vehicles" className="btn btn-primary !px-8 !py-3.5 !text-base shadow-lg shadow-primary-600/25 hover:shadow-primary-600/40">
                <Search size={18} />
                Browse Vehicles
                <ArrowRight size={18} />
              </Link>
              <Link href="/register" className="btn !px-8 !py-3.5 !text-base bg-white/10 text-white hover:bg-white/20 backdrop-blur-sm">
                Create Account
              </Link>
            </div>
            <div className="flex gap-8 mt-12 text-sm text-slate-400">
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-400" />
                200+ Vehicles
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-400" />
                24/7 Support
              </div>
              <div className="flex items-center gap-2">
                <div className="w-2 h-2 rounded-full bg-emerald-400" />
                Best Price Guarantee
              </div>
            </div>
          </div>
        </div>
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-slate-600 to-transparent" />
      </section>

      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-slate-900 mb-4">Why Choose RentCar</h2>
          <p className="text-slate-500 text-lg max-w-2xl mx-auto">
            We make car rental simple, transparent, and enjoyable
          </p>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
          {[
            {
              icon: Car,
              title: "Modern Fleet",
              desc: "Access our carefully curated collection of well-maintained, late-model vehicles for any occasion.",
              color: "bg-primary-100 text-primary-600",
            },
            {
              icon: Calendar,
              title: "Easy Booking",
              desc: "Reserve your vehicle in seconds with our streamlined booking system. Instant confirmation guaranteed.",
              color: "bg-emerald-100 text-emerald-600",
            },
            {
              icon: Search,
              title: "Dynamic Pricing",
              desc: "Enjoy competitive rates that adapt to seasons and demand. Always fair, always transparent.",
              color: "bg-purple-100 text-purple-600",
            },
            {
              icon: Shield,
              title: "Fully Insured",
              desc: "Every rental includes comprehensive insurance coverage. Drive with complete peace of mind.",
              color: "bg-amber-100 text-amber-600",
            },
            {
              icon: Zap,
              title: "Fast Pickup",
              desc: "Skip the lines with our express pickup option. Get on the road in minutes, not hours.",
              color: "bg-rose-100 text-rose-600",
            },
            {
              icon: ArrowRight,
              title: "Flexible Returns",
              desc: "Need to extend? Modify your booking anytime. We adapt to your changing plans.",
              color: "bg-cyan-100 text-cyan-600",
            },
          ].map((feature) => (
            <div key={feature.title} className="card group hover:!shadow-lg hover:-translate-y-1 transition-all duration-300">
              <div className={`w-12 h-12 rounded-xl flex items-center justify-center mb-4 ${feature.color}`}>
                <feature.icon size={24} />
              </div>
              <h3 className="text-lg font-semibold text-slate-900 mb-2 group-hover:text-primary-600 transition-colors">
                {feature.title}
              </h3>
              <p className="text-slate-500 text-sm leading-relaxed">{feature.desc}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="bg-gradient-to-b from-slate-50 to-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
          <div className="bg-gradient-to-br from-primary-600 to-primary-800 rounded-3xl p-8 sm:p-12 lg:p-16 text-white text-center relative overflow-hidden">
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-white/10 via-transparent to-transparent" />
            <div className="relative">
              <h2 className="text-3xl sm:text-4xl font-bold mb-4">Ready to Hit the Road?</h2>
              <p className="text-primary-100 text-lg mb-8 max-w-xl mx-auto">
                Join thousands of satisfied customers who trust RentCar for their journeys.
              </p>
              <div className="flex flex-wrap justify-center gap-4">
                <Link href="/vehicles" className="btn !px-8 !py-3.5 !text-base bg-white text-primary-700 hover:bg-slate-100">
                  Browse Vehicles
                </Link>
                <Link href="/register" className="btn !px-8 !py-3.5 !text-base bg-primary-500/30 text-white hover:bg-primary-500/50 backdrop-blur-sm border border-white/20">
                  Start Free Today
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
