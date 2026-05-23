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
      <main className="min-h-screen bg-background pt-16 text-foreground">
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
