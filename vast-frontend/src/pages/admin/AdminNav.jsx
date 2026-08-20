import { NavLink } from "react-router-dom";
import { LayoutDashboard, Users, ShoppingCart, Server, Box, CreditCard } from "lucide-react";

const links = [
  { to: "/admin", end: true, icon: LayoutDashboard, label: "Overview" },
  { to: "/admin/users", icon: Users, label: "Users" },
  { to: "/admin/offers", icon: ShoppingCart, label: "Offers" },
  { to: "/admin/hosts", icon: Server, label: "Hosts" },
  { to: "/admin/instances", icon: Box, label: "Instances" },
  { to: "/admin/payments", icon: CreditCard, label: "Payments" },
];

export default function AdminNav() {
  return (
    <div className="tabs admin-tabs">
      {links.map(({ to, end, icon: Icon, label }) => (
        <NavLink key={to} to={to} end={end} className={({ isActive }) => `tab${isActive ? " active" : ""}`}>
          <Icon size={14} />
          {label}
        </NavLink>
      ))}
    </div>
  );
}