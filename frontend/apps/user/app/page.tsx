import Navigation from "@/components/landing/Navigation";
import HeroSection from "@/components/landing/HeroSection";
import { SemanticAlertsSection } from "@/components/landing/SemanticAlertsSection";
import { ProblemSection } from "@/components/landing/ProblemSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { CTASection } from "@/components/landing/CTASection";
import Footer from "@/components/landing/Footer";

export default function HomePage() {
  return (
    <>
      <Navigation />
      <main className="min-h-screen bg-[#0a0b0f] bg-[radial-gradient(circle_at_15%_0%,rgba(99,102,241,0.16),transparent_42%),radial-gradient(circle_at_85%_100%,rgba(34,197,94,0.1),transparent_35%)] pt-16 text-slate-100">
        <HeroSection />
        <SemanticAlertsSection />
        <ProblemSection />
        <FeaturesSection />
        <CTASection />
        <Footer />
      </main>
    </>
  );
}
